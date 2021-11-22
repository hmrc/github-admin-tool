package cmd

import (
	"github.com/spf13/cobra"
)

var (
	prApprovalFlag            bool              // nolint // needed for cobra
	prApprovalNumber          int               // nolint // needed for cobra
	prApprovalDismissStale    bool              // nolint // needed for cobra
	prApprovalCodeOwnerReview bool              // nolint // needed for cobra
	prBranchName              string            // nolint // needed for cobra
	prApprovalCmd             = &cobra.Command{ // nolint // needed for cobra
		Use:   "pr-approval",
		Short: "Toggle pr-approval settings for repos in provided list",
		RunE:  prApprovalRun,
	}
)

func prApprovalRun(cmd *cobra.Command, args []string) error {
	approvalArgs := setApprovalArgs(prApprovalCodeOwnerReview, prApprovalDismissStale, prApprovalFlag, prApprovalNumber)
	err := branchProtectionCommand(
		cmd,
		approvalArgs,
		"Pr-approval",
		prBranchName,
		&repository{
			reader: &repositoryReaderService{},
			getter: &repositoryGetterService{},
		},
		&githubRepositorySender{
			sender: &repositorySenderService{},
		},
		&githubBranchProtectionSender{
			sender: &branchProtectionSenderService{},
		},
	)

	return err
}

func setApprovalArgs(
	codeOwnerReview,
	dismissStale,
	approval bool,
	approvalNumber int,
) (branchProtectionArgs []BranchProtectionArgs) {
	return []BranchProtectionArgs{
		{
			Name:     "requiresCodeOwnerReviews",
			DataType: "Boolean",
			Value:    codeOwnerReview,
		},
		{
			Name:     "dismissesStaleReviews",
			DataType: "Boolean",
			Value:    dismissStale,
		},
		{
			Name:     "requiresApprovingReviews",
			DataType: "Boolean",
			Value:    approval,
		},
		{
			Name:     "requiredApprovingReviewCount",
			DataType: "Int",
			Value:    approvalNumber,
		},
	}
}

// nolint // needed for cobra
func init() {
	prApprovalCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "path to file containing repositories (file should contain repos on new line without org/ prefix)")
	prApprovalCmd.Flags().StringVarP(&prBranchName, "branch", "b", "", "branch name to create or update the branch protection rule for")
	prApprovalCmd.Flags().BoolVarP(&prApprovalFlag, "pr-approval", "p", true, "boolean indicating pr reviews before merging, if this is false ignore all other flags")
	prApprovalCmd.Flags().IntVarP(&prApprovalNumber, "number", "n", 1, "number of required approving reviews before PR can be merged")
	prApprovalCmd.Flags().BoolVarP(&prApprovalDismissStale, "dismiss-stale", "d", true, "boolean indicating dismissal of PR review approvals with every new push to branch")
	prApprovalCmd.Flags().BoolVarP(&prApprovalCodeOwnerReview, "code-owner", "o", false, "boolean indicating whether code owner should review")
	prApprovalCmd.MarkFlagRequired("repos")
	prApprovalCmd.MarkFlagRequired("branch")
	prApprovalCmd.Flags().SortFlags = false
	rootCmd.AddCommand(prApprovalCmd)
}
