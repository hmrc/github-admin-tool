package cmd

import (
	"reflect"
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
	var (
		mockDryRun     bool
		mockRepos2File string
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

	type args struct {
		cmd  *cobra.Command
		args []string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "signingRun dry run on",
			args: args{
				cmd: mockCmdWithDryRunOn,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := signingRun(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("signingRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
