package cmd

import (
	"github-admin-tool/graphqlclient"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	prApprovalFlag            bool              // nolint // needed for cobra
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

			repoList, err := repoList(reposFilePath)
			if err != nil {
				log.Fatal(err)
			}

			numberOfRepos := len(repoList)
			if numberOfRepos < 1 || numberOfRepos > maxRepositories {
				log.Fatal("Number of repos passed in must be more than 1 and less than 100")
			}

			if dryRun {
				log.Printf("This is a dry run, the run would process %d repositories", numberOfRepos)

				return
			}

			queryString := repoQuery(repoList)
			client := graphqlclient.NewClient("https://api.github.com/graphql")
			repoSearchResult, err := repoRequest(queryString, client)
			if err != nil {
				log.Fatal(err)
			}
			approvalArgs := setApprovalArgs()
			updated, created, info, problems := branchProtectionApply(
				repoSearchResult,
				"Pr-approval",
				approvalArgs,
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
		},
	}
)

func setApprovalArgs() (branchProtectionArgs []BranchProtectionArgs) {
	branchProtectionArgs = append(
		branchProtectionArgs,
		BranchProtectionArgs{
			Name:     "requiresApprovingReviews",
			DataType: "Boolean",
			Value:    prApprovalFlag,
		},
		{
			Name:     "requiredApprovingReviewCount",
			DataType: "Int",
			Value:    prApprovalNumber,
		},
		{
			Name:     "dismissesStaleReviews",
			DataType: "Boolean",
			Value:    prApprovalDismissStale,
		},
		{
			Name:     "requiresCodeOwnerReviews",
			DataType: "Boolean",
			Value:    prApprovalCodeOwnerReview,
		},
	}
}

// nolint // needed for cobra
func init() {
	prApprovalCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "file containing repositories on new line without org/ prefix. Max 100 repos")
	prApprovalCmd.Flags().BoolVarP(&prApprovalFlag, "pr-approval", "p", true, "boolean indicating pr reviews before merging, if this is false ignore all other flags")
	prApprovalCmd.Flags().IntVarP(&prApprovalNumber, "number", "n", 1, "number of required approving reviews before PR can be merged")
	prApprovalCmd.Flags().BoolVarP(&prApprovalDismissStale, "dismiss-stale", "d", true, "boolean indicating dismissal of PR review approvals with every new push to branch")
	prApprovalCmd.Flags().BoolVarP(&prApprovalCodeOwnerReview, "code-owner", "o", false, "boolean indicating whether code owner should review")
	prApprovalCmd.MarkFlagRequired("repos")
	prApprovalCmd.Flags().SortFlags = false
	rootCmd.AddCommand(prApprovalCmd)
}
