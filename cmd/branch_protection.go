package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var (
	doBranchProtectionSend   = branchProtectionSend   // nolint // Like this for testing mock
	doBranchProtectionUpdate = branchProtectionUpdate // nolint // Like this for testing mock
	doBranchProtectionCreate = branchProtectionCreate // nolint // Like this for testing mock
	doBranchProtectionApply  = branchProtectionApply  // nolint // Like this for testing mock
)

type BranchProtectionArgs struct {
	Name     string
	DataType string
	Value    interface{}
}

func branchProtectionQuery(
	branchProtectionArgs []BranchProtectionArgs,
	action string,
) (
	query string,
	requestVars map[string]interface{},
) {
	mutationBlock, inputBlock, requestVars := branchProtectionQueryBlocks(branchProtectionArgs)

	var mutation, input, output strings.Builder

	mutationName := "createBranchProtectionRule"
	if action == "update" {
		mutationName = "updateBranchProtectionRule"
	}

	mutation.WriteString(fmt.Sprintf("mutation %s(", mutationName))
	mutation.WriteString("$clientMutationId: String!,")

	input.WriteString(fmt.Sprintf("%s(", mutationName))
	input.WriteString("input:{")
	input.WriteString("clientMutationId: $clientMutationId,")

	mutation.WriteString(mutationBlock)
	mutation.WriteString("){")

	input.WriteString(inputBlock)
	input.WriteString("})")

	output.WriteString("{")
	output.WriteString("branchProtectionRule {")
	output.WriteString("id")
	output.WriteString("}")
	output.WriteString("}}")

	query = mutation.String() + input.String() + output.String()

	return query, requestVars
}

func branchProtectionRequest(query string, requestVars map[string]interface{}) *graphqlclient.Request {
	req := graphqlclient.NewRequest(query)
	req.Var("clientMutationId", "github-admin-tool")

	for key, value := range requestVars {
		req.Var(key, value)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	return req
}

func branchProtectionSend(req *graphqlclient.Request) error {
	ctx := context.Background()

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	if err := client.Run(ctx, req, nil); err != nil {
		return fmt.Errorf("from API call: %w", err)
	}

	return nil
}

func branchProtectionQueryBlocks(branchProtectionArgs []BranchProtectionArgs) (
	mutation, input string,
	requestVars map[string]interface{},
) {
	var mutationBlock, inputBlock strings.Builder

	requestVars = make(map[string]interface{})

	for _, bprs := range branchProtectionArgs {
		mutationBlock.WriteString(fmt.Sprintf("$%s: %s!,", bprs.Name, bprs.DataType))
		inputBlock.WriteString(fmt.Sprintf("%s: $%s,", bprs.Name, bprs.Name))
		requestVars[bprs.Name] = bprs.Value
	}

	return mutationBlock.String(), inputBlock.String(), requestVars
}

func branchProtectionApply( // nolint // cyclomatic error 11 !!! Will sort this soon
	repositories map[string]*RepositoriesNode,
	action,
	branchName string,
	branchProtectionArgs []BranchProtectionArgs,
) (
	modified,
	created,
	info,
	problems []string,
) {
	var err error

	for _, repository := range repositories {
		if repository.DefaultBranchRef.Name == "" {
			info = append(info, fmt.Sprintf("No default branch for %v", repository.NameWithOwner))

			continue
		}

		desiredBranchRuleExists := false

		branchProtectionPattern := repository.DefaultBranchRef.Name
		if branchName != "" {
			branchProtectionPattern = branchName
		}

		// Check all nodes for default branch protection rule
		for _, branchProtection := range repository.BranchProtectionRules.Nodes {
			if branchProtectionPattern == branchProtection.Pattern {
				desiredBranchRuleExists = true
			}

			updateRequired, returnInfo := branchProtectionUpdateCheck(action, branchProtectionPattern, branchProtection)
			if returnInfo {
				info = append(
					info,
					fmt.Sprintf(
						"%s already turned on for %v with branch name: %s",
						action,
						repository.NameWithOwner,
						branchProtection.Pattern,
					),
				)

				continue
			}

			if updateRequired {
				if err = doBranchProtectionUpdate(branchProtectionArgs, branchProtection.ID); err != nil {
					problems = append(problems, err.Error())

					continue
				}

				modified = append(
					modified,
					fmt.Sprintf(
						"%s changed for %v with branch name: %s",
						action,
						repository.NameWithOwner,
						branchProtection.Pattern,
					),
				)

				continue
			}
		}

		if !desiredBranchRuleExists {
			if err = doBranchProtectionCreate(
				branchProtectionArgs,
				repository.ID,
				branchProtectionPattern,
			); err != nil {
				problems = append(problems, err.Error())

				continue
			}

			created = append(
				created,
				fmt.Sprintf(
					"Branch protection rule created for %v with branch name: %s",
					repository.NameWithOwner,
					branchProtectionPattern,
				),
			)
		}
	}

	return modified, created, info, problems
}

func branchProtectionUpdateCheck(
	action,
	branchNamePattern string,
	branchProtection BranchProtectionRulesNode,
) (
	updateRequired bool,
	returnInfo bool,
) {
	// If default branch has already got signing turned on, no need to update
	if action == "Signing" {
		if branchProtection.RequiresCommitSignatures {
			return false, true
		}
	}

	// If rule pattern doesn't match branch flag then ignore update
	if action == "Pr-approval" {
		if branchProtection.Pattern != branchNamePattern {
			return false, false
		}

		// If default branch has already got pr-approval turned on, no need to update
		if branchProtection.RequiresApprovingReviews == prApprovalFlag &&
			branchProtection.RequiredApprovingReviewCount == prApprovalNumber &&
			branchProtection.DismissesStaleReviews == prApprovalDismissStale &&
			branchProtection.RequiresCodeOwnerReviews == prApprovalCodeOwnerReview {
			return false, true
		}
	}

	return true, false
}

func branchProtectionUpdate(branchProtectionArgs []BranchProtectionArgs, branchProtectionRuleID string) error {
	branchProtectionArgs = append(
		branchProtectionArgs,
		BranchProtectionArgs{
			Name:     "branchProtectionRuleId",
			DataType: "String",
			Value:    branchProtectionRuleID,
		},
	)
	query, requestVars := branchProtectionQuery(branchProtectionArgs, "update")
	req := branchProtectionRequest(query, requestVars)
	err := doBranchProtectionSend(req)

	return err
}

func branchProtectionCreate(branchProtectionArgs []BranchProtectionArgs, repositoryID, pattern string) error {
	branchProtectionArgs = append(
		branchProtectionArgs,
		BranchProtectionArgs{
			Name:     "repositoryId",
			DataType: "String",
			Value:    repositoryID,
		},
		BranchProtectionArgs{
			Name:     "pattern",
			DataType: "String",
			Value:    pattern,
		},
	)
	query, requestVars := branchProtectionQuery(branchProtectionArgs, "create")
	req := branchProtectionRequest(query, requestVars)
	err := doBranchProtectionSend(req)

	return err
}

func branchProtectionCommand(
	cmd *cobra.Command,
	branchProtectionArgs []BranchProtectionArgs,
	action,
	branchName string,
) error {
	dryRun, reposFilePath, err := branchProtectionFlagCheck(cmd)
	if err != nil {
		return err
	}

	repositoryList, err := branchProtectionRepoList(reposFilePath, 100)
	if err != nil {
		return err
	}

	log.SetFlags(0)

	if dryRun {
		log.Printf("This is a dry run, the run would process %d repositories", len(repositoryList))

		return nil
	}

	repositories, err := doRepositoryGet(repositoryList)
	if err != nil {
		return err
	}

	updated, created, info, problems := doBranchProtectionApply(
		repositories,
		action,
		branchName,
		branchProtectionArgs,
	)

	for key, repo := range updated {
		log.Printf("Modified (%d): %v", key, repo)
	}

	for key, repo := range created {
		log.Printf("Created (%d): %v", key, repo)
	}

	for key, err := range problems {
		log.Printf("Error (%d): %v", key, err)
	}

	for key, i := range info {
		log.Printf("Info (%d): %v", key, i)
	}

	return nil
}

func branchProtectionFlagCheck(cmd *cobra.Command) (dryRun bool, reposFilePath string, err error) {
	dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return dryRun, reposFilePath, fmt.Errorf("%w", err)
	}

	reposFilePath, err = cmd.Flags().GetString("repos")
	if err != nil {
		return dryRun, reposFilePath, fmt.Errorf("%w", err)
	}

	return dryRun, reposFilePath, nil
}

func branchProtectionRepoList(reposFilePath string, maxRepositories int) ([]string, error) {
	repositoryList, err := repositoryList(reposFilePath)
	if err != nil {
		return []string{}, err
	}

	numberOfRepos := len(repositoryList)
	if numberOfRepos < 1 || numberOfRepos > maxRepositories {
		return []string{}, errTooManyRepos
	}

	return repositoryList, nil
}
