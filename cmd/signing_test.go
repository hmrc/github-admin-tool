package cmd

import (
	"reflect"
	"testing"
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
