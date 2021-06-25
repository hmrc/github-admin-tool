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
			name: "prApprovalRun success",
			args: args{
				cmd: mockCmd,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: dry-run",
		},
		{
			name: "prApprovalRun success",
			args: args{
				cmd: mockCmdWithDryRun,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: dry-run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := prApprovalRun(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("prApprovalRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
