package cmd

import (
	"bytes"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func Test_setSigningArgs(t *testing.T) {
	tests := []struct {
		name                     string
		wantBranchProtectionArgs []BranchProtectionArgs
	}{
		{
			name: "set SigningArgs return values are as expected",
			wantBranchProtectionArgs: []BranchProtectionArgs{
				{
					Name:     "requiresCommitSignatures",
					DataType: "Boolean",
					Value:    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBranchProtectionArgs := setSigningArgs(); !reflect.DeepEqual(
				gotBranchProtectionArgs,
				tt.wantBranchProtectionArgs,
			) {
				t.Errorf("setSigningArgs() = %v, want %v", gotBranchProtectionArgs, tt.wantBranchProtectionArgs)
			}
		})
	}
}

func Test_signingRun(t *testing.T) {
	doRepositoryGet = mockRepositoryGet
	defer func() { doRepositoryGet = repositoryGet }()

	doBranchProtectionApply = mockBranchProtectionApply
	defer func() { doBranchProtectionApply = branchProtectionApply }()

	mockCmd := &cobra.Command{
		Use: "signing",
	}

	mockCmdWithDryRun := &cobra.Command{
		Use: "signing",
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
		Use: "signing",
	}
	mockCmdWithDryRunAndRepos.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunAndRepos.Flags().StringVarP(&mockReposFile, "repos", "r", "", "repos file")

	mockCmdWithDryRunAndTooManyRepos := &cobra.Command{
		Use: "signing",
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
		Use: "signing",
	}
	mockCmdWithDryRunOn.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunOn.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"../testdata/two_repo_list.txt",
		"repos file",
	)

	mockCmdWithDryRunOff := &cobra.Command{
		Use: "signing",
	}
	mockCmdWithDryRunOff.Flags().BoolVarP(&mockDryRunFalse, "dry-run", "d", false, "dry run flag")
	mockCmdWithDryRunOff.Flags().StringVarP(
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
		name              string
		args              args
		wantErr           bool
		wantErrMsg        string
		mockErrorFunction bool
		wantLogOutput     string
	}{
		{
			name: "signinglRun fails dry run",
			args: args{
				cmd: mockCmd,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: dry-run",
		},
		{
			name: "signingRun fails repo",
			args: args{
				cmd: mockCmdWithDryRun,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: repos",
		},
		{
			name: "signingRun fails repo list path",
			args: args{
				cmd: mockCmdWithDryRunAndRepos,
			},
			wantErr:    true,
			wantErrMsg: "could not open repo file: open : no such file or directory",
		},
		{
			name: "signingRun number of repos too big",
			args: args{
				cmd: mockCmdWithDryRunAndTooManyRepos,
			},
			wantErr:    true,
			wantErrMsg: "number of repos passed in must be more than 1 and less than 100",
		},
		{
			name: "signingRun dry run on",
			args: args{
				cmd: mockCmdWithDryRunOn,
			},
			wantErr:       false,
			wantLogOutput: "This is a dry run, the run would process 2 repositories",
		},
		{
			name: "signingRun repo get fails",
			args: args{
				cmd: mockCmdWithDryRunOff,
			},
			wantErr:           true,
			wantErrMsg:        "repo failure",
			mockErrorFunction: true,
		},

		{
			name: "signingRun check log output",
			args: args{
				cmd: mockCmdWithDryRunOff,
			},
			wantErr: false,
			wantLogOutput: `Modified (0): modified branch
Created (0): created branch
Error (0): problems branch
Info (0): info branch`,
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

			err := signingRun(tt.args.cmd, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("signingRun() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("signingRun() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}

			if tt.wantLogOutput != strings.TrimSpace(buf.String()) {
				t.Errorf("signingRun() log = \n\n%v, wantLogOutput \n\n%v", buf.String(), tt.wantLogOutput)
			}
		})
	}
}
