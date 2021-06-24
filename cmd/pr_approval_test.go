package cmd

import (
	"reflect"
	"testing"
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
