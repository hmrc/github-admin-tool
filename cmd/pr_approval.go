package cmd

import (
	"errors"

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
		Short: "Set request signing on to all repos in provided list",
		RunE:  prApprovalRun,
	}
	errTooManyRepos = errors.New("number of repos passed in must be more than 1 and less than 100")
)

func prApprovalRun(cmd *cobra.Command, args []string) error {
	approvalArgs := setApprovalArgs()
	err := branchProtectionCommand(cmd, approvalArgs, "Pr-approval", prBranchName)

	return err
}

func setApprovalArgs() (branchProtectionArgs []BranchProtectionArgs) {
	return []BranchProtectionArgs{
		{
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
