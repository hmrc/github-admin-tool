package cmd

import (
	"bytes"
	"errors"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func Test_setApprovalArgs(t *testing.T) {
	type args struct {
		codeOwnerReview bool
		dismissStale    bool
		approval        bool
		approvalNumber  int
	}

	tests := []struct {
		name                     string
		args                     args
		wantBranchProtectionArgs []BranchProtectionArgs
	}{
		{
			name: "setApprovalargs returned values are as expected",
			args: args{
				codeOwnerReview: false,
				dismissStale:    true,
				approval:        true,
				approvalNumber:  5,
			},
			wantBranchProtectionArgs: []BranchProtectionArgs{
				{
					Name:     "requiresCodeOwnerReviews",
					DataType: "Boolean",
					Value:    false,
				},
				{
					Name:     "dismissesStaleReviews",
					DataType: "Boolean",
					Value:    true,
				},
				{
					Name:     "requiresApprovingReviews",
					DataType: "Boolean",
					Value:    true,
				},
				{
					Name:     "requiredApprovingReviewCount",
					DataType: "Int",
					Value:    5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBranchProtectionArgs := setApprovalArgs(
				tt.args.codeOwnerReview,
				tt.args.dismissStale,
				tt.args.approval,
				tt.args.approvalNumber,
			); !reflect.DeepEqual(
				gotBranchProtectionArgs,
				tt.wantBranchProtectionArgs,
			) {
				t.Errorf("setApprovalArgs() = %v, want %v", gotBranchProtectionArgs, tt.wantBranchProtectionArgs)
			}
		})
	}
}

var errRepositoryGet = errors.New("repo failure")

func mockRepositoryGet([]string) (repoNode map[string]*RepositoriesNode, err error) {
	return repoNode, nil
}

func mockRepositoryGetError([]string) (repoNode map[string]*RepositoriesNode, err error) {
	return repoNode, errRepositoryGet
}

func mockBranchProtectionApply(
	repositories map[string]*RepositoriesNode,
	action, branchName string,
	branchProtectionArgs []BranchProtectionArgs,
) (
	updated,
	created,
	info,
	problems []string,
) {
	return []string{"updated branch"},
		[]string{"created branch"},
		[]string{"info branch"},
		[]string{"problems branch"}
}

func Test_prApprovalRun(t *testing.T) {
	doRepositoryGet = mockRepositoryGet
	defer func() { doRepositoryGet = repositoryGet }()

	doBranchProtectionApply = mockBranchProtectionApply
	defer func() { doBranchProtectionApply = branchProtectionApply }()

	mockCmd := &cobra.Command{
		Use: "pr-approval",
	}

	mockCmdWithDryRun := &cobra.Command{
		Use: "pr-approval",
	}

	var (
		mockDryRun           bool
		mockDryRunFalse      bool
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
		"testdata/repo_list_more_than_hundred.txt",
		"repos file",
	)

	mockCmdWithDryRunOn := &cobra.Command{
		Use: "pr-approval",
	}
	mockCmdWithDryRunOn.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunOn.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"testdata/two_repo_list.txt",
		"repos file",
	)

	mockCmdWithDryRunOff := &cobra.Command{
		Use: "pr-approval",
	}
	mockCmdWithDryRunOff.Flags().BoolVarP(&mockDryRunFalse, "dry-run", "d", false, "dry run flag")
	mockCmdWithDryRunOff.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"testdata/two_repo_list.txt",
		"repos file",
	)

	type args struct {
		cmd  *cobra.Command
		args []string
	}

	tests := []struct {
		name              string
		args              args
		wantErr           bool
		wantErrMsg        string
		mockErrorFunction bool
		wantLogOutput     string
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
			name: "prApprovalRun dry run on",
			args: args{
				cmd: mockCmdWithDryRunOn,
			},
			wantErr:       false,
			wantLogOutput: "This is a dry run, the run would process 2 repositories",
		},
		{
			name: "prApprovalRun repo get fails",
			args: args{
				cmd: mockCmdWithDryRunOff,
			},
			wantErr:           true,
			wantErrMsg:        "repo failure",
			mockErrorFunction: true,
		},

		{
			name: "prApprovalRun check log output",
			args: args{
				cmd: mockCmdWithDryRunOff,
			},
			wantErr: false,
			wantLogOutput: `Updated (Batch 0-2): updated branch
Created (Batch 0-2): created branch
Info (Batch 0-2): info branch
Error (Batch 0-2): problems branch`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErrorFunction {
				doRepositoryGet = mockRepositoryGetError
				defer func() { doRepositoryGet = mockRepositoryGet }()
			}

			var buf bytes.Buffer

			log.SetOutput(&buf)

			defer func() { log.SetOutput(os.Stderr) }()

			err := prApprovalRun(tt.args.cmd, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("prApprovalRun() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("prApprovalRun() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}

			if tt.wantLogOutput != strings.TrimSpace(buf.String()) {
				t.Errorf("prApprovalRun() log = \n\n%v, wantLogOutput \n\n%v", buf.String(), tt.wantLogOutput)
			}
		})
	}
}
