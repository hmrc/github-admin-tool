package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	prApprovalNumber          int               // nolint // needed for cobra
	prApprovalDismissStale    bool              // nolint // needed for cobra
	prApprovalCodeOwnerReview bool              // nolint // needed for cobra
	prApprovalCmd             = &cobra.Command{ // nolint // needed for cobra
		Use:   "pr-approval",
		Short: "Set request signing on to all repos in provided list",
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, err := cmd.Flags().GetBool("dry-run")
			if err != nil {
				log.Fatal(err)
			}

			reposFilePath, err := cmd.Flags().GetString("repos")
			if err != nil {
				log.Fatal(err)
			}

			repoList, err := readRepoList(reposFilePath)
			if err != nil {
				log.Fatal(err)
			}

			numberOfRepos := len(repoList)
			if numberOfRepos < 1 || numberOfRepos > maxRepositories {
				log.Fatal("Number of repos passed in must be more than 1 and less than 100")
			}

			if dryRun {
				log.Printf("This is a dry run, the run would process %d repositories", numberOfRepos)
				os.Exit(0)
			}

			queryString := generateRepoQuery(repoList)
			client := graphqlclient.NewClient("https://api.github.com/graphql")
			repoSearchResult, err := repoRequest(queryString, client)
			if err != nil {
				log.Fatal(err)
			}
			updated, created, info, problems := applyPrApproval(repoSearchResult, client)

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
		},
	}
)

func applyPrApproval(repoSearchResult map[string]RepositoriesNodeList, client *graphqlclient.Client) (
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
		for _, branchProtectionRule := range repository.BranchProtectionRules.Nodes {
			if repository.DefaultBranchRef.Name != branchProtectionRule.Pattern {
				continue
			}

			updateArgs := setApprovalArgs(branchProtectionRule.ID)
			if err = prApprovalUpdate(branchProtectionRule.ID, updateArgs, client); err != nil {
				problems = append(problems, err.Error())

				continue OUTER
			}
			modified = append(modified, repository.NameWithOwner)

			continue OUTER
		}

		if err = prApprovalCreate(repository.ID, repository.DefaultBranchRef.Name, client); err != nil {
			problems = append(problems, err.Error())

			continue OUTER
		}

		created = append(created, repository.NameWithOwner)
	}

	return modified, created, info, problems
}

func setApprovalArgs(branchProtectionRuleID string) (branchProtectionArgs []BranchProtectionArgs) {
	branchProtectionArgs = append(branchProtectionArgs, BranchProtectionArgs{
		Name:     "requiresApprovingReviews",
		DataType: "Boolean",
		Value:    true,
	})

	branchProtectionArgs = append(branchProtectionArgs, BranchProtectionArgs{
		Name:     "requiredApprovingReviewCount",
		DataType: "Int",
		Value:    prApprovalNumber,
	})

	branchProtectionArgs = append(branchProtectionArgs, BranchProtectionArgs{
		Name:     "dismissesStaleReviews",
		DataType: "Boolean",
		Value:    prApprovalDismissStale,
	})

	branchProtectionArgs = append(branchProtectionArgs, BranchProtectionArgs{
		Name:     "requiresCodeOwnerReviews",
		DataType: "Boolean",
		Value:    prApprovalCodeOwnerReview,
	})

	return branchProtectionArgs
}

func createPrApprovalBranchProtection(repositoryID, branchName string, client *graphqlclient.Client) error {
	req := graphqlclient.NewRequest(`
		mutation CreateBranchProtectionRule(
			$repositoryId: String!,
			$clientMutationId: String!,
			$requiredApprovingReviewCount: Int!,
			$pattern: String!,
			$dismissesStaleReviews: Boolean!,
			$requiresCodeOwnerReviews: Boolean!
		) {
			createBranchProtectionRule(
				input:{
					clientMutationId: $clientMutationId,
					repositoryId: $repositoryId,
					requiresApprovingReviews: true, 
					requiredApprovingReviewCount: $requiredApprovingReviewCount,
					pattern: $pattern,
					dismissesStaleReviews:  $dismissesStaleReviews,
					requiresCodeOwnerReviews: $requiresCodeOwnerReviews
				}
			) {
				branchProtectionRule {
					id
				}
			}
		}
	`)

	req.Var("clientMutationId", fmt.Sprintf("github-tool-%v", repositoryID))
	req.Var("repositoryId", repositoryID)
	req.Var("requiredApprovingReviewCount", prApprovalNumber)
	req.Var("pattern", branchName)
	req.Var("dismissesStaleReviews", prApprovalDismissStale)
	req.Var("requiresCodeOwnerReviews", prApprovalCodeOwnerReview)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	ctx := context.Background()

	if err := client.Run(ctx, req, nil); err != nil {
		return fmt.Errorf("from API call: %w", err)
	}

	return nil
}

// nolint // needed for cobra
func init() {
	prApprovalCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "file containing repositories on new line without org/ prefix. Max 100 repos")
	prApprovalCmd.Flags().BoolVarP(&prApprovalCodeOwnerReview, "code-owner", "o", false, "boolean indicating whether code owner should review")
	prApprovalCmd.Flags().IntVarP(&prApprovalNumber, "number", "n", 1, "number of required approving reviews before PR can be merged")
	prApprovalCmd.Flags().BoolVarP(&prApprovalDismissStale, "dismiss-stale", "d", true, "boolean indicating dismissal of PR review approvals with every new push to branch")
	prApprovalCmd.MarkFlagRequired("repos")
	rootCmd.AddCommand(prApprovalCmd)
}
