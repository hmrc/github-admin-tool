package cmd

import (
	"reflect"
	"testing"
)

var TestEmptyCSVRows = [][]string{ // nolint // want to use this in both tests
	{
		"Repo Name",
		"Default Branch Name",
		"Is Archived",
		"Is Private",
		"Is Empty",
		"Is Fork",
		"Parent Repo Name",
		"Merge Commit Allowed",
		"Squash Merge Allowed",
		"Rebase Merge Allowed",
		"(BP1) IsAdminEnforced",
		"(BP1) RequiresCommitSignatures",
		"(BP1) RestrictsPushes",
		"(BP1) RequiresApprovingReviews",
		"(BP1) RequiresStatusChecks",
		"(BP1) RequiresCodeOwnerReviews",
		"(BP1) DismissesStaleReviews",
		"(BP1) RequiresStrictStatusChecks",
		"(BP1) RequiredApprovingReviewCount",
		"(BP1) AllowsForcePushes",
		"(BP1) AllowsDeletions",
		"(BP1) Branch Protection Pattern",
		"(BP2) IsAdminEnforced",
		"(BP2) RequiresCommitSignatures",
		"(BP2) RestrictsPushes",
		"(BP2) RequiresApprovingReviews",
		"(BP2) RequiresStatusChecks",
		"(BP2) RequiresCodeOwnerReviews",
		"(BP2) DismissesStaleReviews",
		"(BP2) RequiresStrictStatusChecks",
		"(BP2) RequiredApprovingReviewCount",
		"(BP2) AllowsForcePushes",
		"(BP2) AllowsDeletions",
		"(BP2) Branch Protection Pattern",
	},
}

func Test_parse(t *testing.T) {
	var (
		emptyAllResults []ReportResponse
		TestEmptyList   [][]string
	)

	type args struct {
		ignoreArchived bool
		allResults     []ReportResponse
	}

	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "Parse empty list return empty",
			args: args{ignoreArchived: false, allResults: emptyAllResults},
			want: TestEmptyList,
		},
		{
			name: "Parse archived result set",
			args: args{ignoreArchived: true, allResults: []ReportResponse{{
				Organization{Repositories{Nodes: []RepositoriesNode{{IsArchived: true}}}},
			}}},
			want: TestEmptyList,
		},
		{
			name: "Parse unarchived result set",
			args: args{ignoreArchived: true, allResults: []ReportResponse{{
				Organization{Repositories{Nodes: []RepositoriesNode{{IsArchived: false, NameWithOwner: "REPONAME1"}}}},
			}}},
			want: [][]string{{"REPONAME1", "", "false", "false", "false", "false", "", "false", "false", "false"}},
		},
		{
			name: "Parse branch protection result set",
			args: args{
				ignoreArchived: true,
				allResults: []ReportResponse{{
					Organization{
						Repositories{
							Nodes: []RepositoriesNode{{
								BranchProtectionRules: BranchProtectionRules{
									Nodes: []BranchProtectionRulesNode{{
										Pattern: "SOMEREGEXP",
									}},
								},
							}},
						},
					},
				}},
			},
			want: [][]string{{
				"", "", "false", "false", "false", "false", "", "false", "false", "false", "false",
				"false", "false", "false", "false", "false", "false", "false", "0", "false", "false", "SOMEREGEXP",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parse(tt.args.ignoreArchived, tt.args.allResults); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeCSV(t *testing.T) {
	var TestEmptyList [][]string

	wantWithBP := make([][]string, 1)
	wantWithBP[0] = append(
		wantWithBP[0],
		"REPONAME1", "", "false", "false", "false", "false", "", "false", "false",
		"false", "false", "false", "false", "false", "false", "false", "false",
		"false", "0", "false", "false", "SOMEREGEXP",
	)

	twoCSVRows := TestEmptyCSVRows
	twoCSVRows = append(twoCSVRows, wantWithBP...)

	type args struct {
		parsed [][]string
	}

	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "WriteCSV returns no extra rows",
			args: args{parsed: TestEmptyList},
			want: TestEmptyCSVRows,
		},
		{
			name: "WriteCSV returns some rows",
			args: args{parsed: wantWithBP},
			want: twoCSVRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeCSV(tt.args.parsed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("writeCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}
