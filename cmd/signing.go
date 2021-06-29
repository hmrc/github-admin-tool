package cmd

import (
	"github.com/spf13/cobra"
)

var (
	requiresCommitSignatures bool              // nolint // needed for cobra
	signingCmd               = &cobra.Command{ // nolint // needed for cobra
		Use:   "signing",
		Short: "Set request signing on to all repos in provided list",
		RunE:  signingRun,
	}
)

func signingRun(cmd *cobra.Command, args []string) error {
	signingArgs := setSigningArgs()
	err := branchProtectionCommand(cmd, signingArgs, "Signing", "")

	return err
}

func setSigningArgs() (branchProtectionArgs []BranchProtectionArgs) {
	branchProtectionArgs = append(
		branchProtectionArgs,
		BranchProtectionArgs{
			Name:     "requiresCommitSignatures",
			DataType: "Boolean",
			Value:    true,
		})

	return branchProtectionArgs
}

// nolint // needed for cobra
func init() {
	signingCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "file containing repositories on new line without org/ prefix. Max 100 repos")
	signingCmd.MarkFlagRequired("repos")
	rootCmd.AddCommand(signingCmd)
}
