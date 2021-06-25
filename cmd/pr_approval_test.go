package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func Test_setApprovalArgs(t *testing.T) {
	tests := []struct {
		name                     string
		wantBranchProtectionArgs []BranchProtectionArgs
	}{
		{
			name: "setApprovalargs returned values are as expected",
			wantBranchProtectionArgs: []BranchProtectionArgs{
				{
					Name:     "requiresApprovingReviews",
					DataType: "Boolean",
					Value:    true,
				},
				{
					Name:     "requiredApprovingReviewCount",
					DataType: "Int",
					Value:    1,
				},
				{
					Name:     "dismissesStaleReviews",
					DataType: "Boolean",
					Value:    true,
				},
				{
					Name:     "requiresCodeOwnerReviews",
					DataType: "Boolean",
					Value:    false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBranchProtectionArgs := setApprovalArgs(); !reflect.DeepEqual(
				gotBranchProtectionArgs,
				tt.wantBranchProtectionArgs,
			) {
				t.Errorf("setApprovalArgs() = %v, want %v", gotBranchProtectionArgs, tt.wantBranchProtectionArgs)
			}
		})
	}
}

func Test_prApprovalRun(t *testing.T) {
	mockCmd := &cobra.Command{
		Use: "pr-approval",
	}

	mockCmdWithDryRun := &cobra.Command{
		Use: "pr-approval",
	}

	var (
		mockDryRun           bool
		mockReposFile        string
		mockReposOver100File string
		mockRepos2File       string
	)

	mockCmdWithDryRun.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")

	mockCmdWithDryRunAndRepos := &cobra.Command{
		Use: "pr-approval",
	}
	mockCmdWithDryRunAndRepos.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunAndRepos.Flags().StringVarP(&mockReposFile, "repos", "r", "", "repos file")

	mockCmdWithDryRunAndTooManyRepos := &cobra.Command{
		Use: "pr-approval",
	}
	mockCmdWithDryRunAndTooManyRepos.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunAndTooManyRepos.Flags().StringVarP(
		&mockReposOver100File,
		"repos",
		"r",
		"../testdata/repo_list_more_than_hundred.txt",
		"repos file",
	)

	mockCmdWithDryRunOn := &cobra.Command{
		Use: "pr-approval",
	}
	mockCmdWithDryRunOn.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdWithDryRunOn.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"../testdata/two_repo_list.txt",
		"repos file",
	)

	type args struct {
		cmd  *cobra.Command
		args []string
	}

	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "prApprovalRun fails dry run",
			args: args{
				cmd: mockCmd,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: dry-run",
		},
		{
			name: "prApprovalRun fails repo",
			args: args{
				cmd: mockCmdWithDryRun,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: repos",
		},
		{
			name: "prApprovalRun fails repo list path",
			args: args{
				cmd: mockCmdWithDryRunAndRepos,
			},
			wantErr:    true,
			wantErrMsg: "could not open repo file: open : no such file or directory",
		},
		{
			name: "prApprovalRun number of repos too big",
			args: args{
				cmd: mockCmdWithDryRunAndTooManyRepos,
			},
			wantErr:    true,
			wantErrMsg: "number of repos passed in must be more than 1 and less than 100",
		},
		{
			name: "prApprovalRun dry run on",
			args: args{
				cmd: mockCmdWithDryRunOn,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := prApprovalRun(tt.args.cmd, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("prApprovalRun() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("prApprovalRun() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}
