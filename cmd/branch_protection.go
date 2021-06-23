package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"strings"
)

func branchProtectionQuery(branchProtectionArgs []BranchProtectionArgs, action string) (string, map[string]interface{}) {
	mutationBlock, inputBlock, requestVars := branchProtectionQueryBlocks(branchProtectionArgs)

	var mutation, input, output strings.Builder

	mutationName := "createBranchProtectionRule"
	if action == "update" {
		mutationName = "updateBranchProtectionRule"
	}

	mutation.WriteString(fmt.Sprintf("	mutation %s(", mutationName))
	mutation.WriteString("		$clientMutationId: String!,")

	input.WriteString(fmt.Sprintf("	%s(", mutationName))
	input.WriteString("		input:{")
	input.WriteString("			clientMutationId: $clientMutationId,")

	mutation.WriteString(mutationBlock)
	mutation.WriteString("){")

	input.WriteString(inputBlock)
	input.WriteString("})")

	output.WriteString("{")
	output.WriteString("	branchProtectionRule {")
	output.WriteString("		id")
	output.WriteString("	}")
	output.WriteString("}}")

	return mutation.String() + input.String() + output.String(), requestVars
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

func branchProtectionSend(req *graphqlclient.Request, client *graphqlclient.Client) error {
	ctx := context.Background()

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

func branchProtectionApply(
	repoSearchResult map[string]RepositoriesNode,
	action string,
	branchProtectionArgs []BranchProtectionArgs,
	client *graphqlclient.Client,
) (
	modified,
	created,
	info,
	problems []string,
) {
	var err error

OUTER:

	for _, repository := range repoSearchResult { // nolint
		if repository.DefaultBranchRef.Name == "" {
			info = append(info, fmt.Sprintf("No default branch for %v", repository.NameWithOwner))

			continue OUTER
		}

		// Check all nodes for default branch protection rule
		for _, branchProtection := range repository.BranchProtectionRules.Nodes {
			if repository.DefaultBranchRef.Name != branchProtection.Pattern {
				continue
			}

			// If default branch has already got signing turned on, no need to update
			if action == "Signing" && branchProtection.RequiresCommitSignatures {
				info = append(info, fmt.Sprintf("%s already turned on for %v", action, repository.NameWithOwner))

				continue OUTER
			}

			if err = branchProtectionUpdate(branchProtectionArgs, branchProtection.ID); err != nil {
				problems = append(problems, err.Error())

				continue OUTER
			}
			modified = append(modified, repository.NameWithOwner)

			continue OUTER
		}

		if err = branchProtectionCreate(
			branchProtectionArgs,
			repository.ID,
			repository.DefaultBranchRef.Name,
		); err != nil {
			problems = append(problems, err.Error())

			continue OUTER
		}

		created = append(created, repository.NameWithOwner)
	}

	return modified, created, info, problems
}

func branchProtectionUpdate(branchProtectionArgs []BranchProtectionArgs, branchProtectionRuleId string) error {
	branchProtectionArgs = append(
		branchProtectionArgs,
		BranchProtectionArgs{
			Name:     "branchProtectionRuleId",
			DataType: "String",
			Value:    branchProtectionRuleId,
		},
	)
	query, requestVars := branchProtectionQuery(branchProtectionArgs, "update")
	req := branchProtectionRequest(query, requestVars)
	client := graphqlclient.NewClient("https://api.github.com/graphql")
	if err := doBranchProtectionSend(req, client); err != nil {
		return err
	}

	return nil
}

func branchProtectionCreate(branchProtectionArgs []BranchProtectionArgs, repoId string, pattern string) error {
	branchProtectionArgs = append(
		branchProtectionArgs,
		BranchProtectionArgs{
			Name:     "repositoryId",
			DataType: "String",
			Value:    repoId,
		},
		BranchProtectionArgs{
			Name:     "pattern",
			DataType: "String",
			Value:    pattern,
		},
	)
	query, requestVars := branchProtectionQuery(branchProtectionArgs, "create")
	req := branchProtectionRequest(query, requestVars)
	client := graphqlclient.NewClient("https://api.github.com/graphql")
	if err := doBranchProtectionSend(req, client); err != nil {
		return err
	}

	return nil
}
