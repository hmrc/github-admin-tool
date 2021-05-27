package cmd

import (
	"reflect"
	"testing"
)

var TestEmptyList [][]string

var TestEmptyCsvRows = [][]string{
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
	var emptyAllResults []ReportResponse

	responsesWithArchived := make([]ReportResponse, 1)
	responsesWithArchived[0].Organization.Repositories.Nodes = append(responsesWithArchived[0].Organization.Repositories.Nodes, RepositoriesNodeList{IsArchived: true})

	responsesWithUnarchived := make([]ReportResponse, 1)
	responsesWithUnarchived[0].Organization.Repositories.Nodes = append(responsesWithUnarchived[0].Organization.Repositories.Nodes, RepositoriesNodeList{IsArchived: false, NameWithOwner: "REPONAME1"})

	wantWithUnarchived := make([][]string, 1)
	wantWithUnarchived[0] = append(wantWithUnarchived[0], "REPONAME1", "", "false", "false", "false", "false", "", "false", "false", "false")

	branchProtectionList := make([]BranchProtectionRulesNodesList, 1)
	branchProtectionList[0].Pattern = " SOMEREGEXP "
	repositoriesNodeList := make([]RepositoriesNodeList, 1)
	repositoriesNodeList[0].BranchProtectionRules.Nodes = branchProtectionList
	responsesWithBP := make([]ReportResponse, 1)
	responsesWithBP[0].Organization.Repositories.Nodes = repositoriesNodeList

	wantWithBP := make([][]string, 1)
	wantWithBP[0] = append(wantWithBP[0], "", "", "false", "false", "false", "false", "", "false", "false", "false", "false", "false", "false", "false", "false", "false", "false", "false", "0", "false", "false", "SOMEREGEXP")

	type args struct {
		ignoreArchived bool
		allResults     []ReportResponse
	}
	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{name: "ParseEmptyListReturnEmpty", args: args{ignoreArchived: false, allResults: emptyAllResults}, want: TestEmptyList},
		{name: "ParseEmptyList", args: args{ignoreArchived: false}, want: TestEmptyList},
		{name: "ParseArchivedResultSet", args: args{ignoreArchived: true, allResults: responsesWithArchived}, want: TestEmptyList},
		{name: "ParseUnarchivedResultSet", args: args{ignoreArchived: true, allResults: responsesWithUnarchived}, want: wantWithUnarchived},
		{name: "ParseBranchProtectionResultSet", args: args{ignoreArchived: true, allResults: responsesWithBP}, want: wantWithBP},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parse(tt.args.ignoreArchived, tt.args.allResults); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeCsv(t *testing.T) {
	wantWithBP := make([][]string, 1)
	wantWithBP[0] = append(wantWithBP[0], "REPONAME1", "", "false", "false", "false", "false", "", "false", "false", "false", "false", "false", "false", "false", "false", "false", "false", "false", "0", "false", "false", "SOMEREGEXP")

	twoCsvRows := TestEmptyCsvRows
	twoCsvRows = append(twoCsvRows, wantWithBP...)

	type args struct {
		parsed [][]string
	}
	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{name: "WriteCSVReturnsNoExtraRows", args: args{parsed: TestEmptyList}, want: TestEmptyCsvRows},
		{name: "WriteCSVReturnsSomeRows", args: args{parsed: wantWithBP}, want: twoCsvRows},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeCsv(tt.args.parsed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("writeCsv() = %v, want %v", got, tt.want)
			}
		})
	}
}
