package cmd

import (
	"errors"
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

func Test_reportCSVParse(t *testing.T) {
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
			name: "reportCSVParse empty list return empty",
			args: args{ignoreArchived: false, allResults: emptyAllResults},
			want: TestEmptyList,
		},
		{
			name: "reportCSVParse archived result set",
			args: args{ignoreArchived: true, allResults: []ReportResponse{{
				Organization{Repositories{Nodes: []RepositoriesNode{{IsArchived: true}}}},
			}}},
			want: TestEmptyList,
		},
		{
			name: "reportCSVParse unarchived result set",
			args: args{ignoreArchived: true, allResults: []ReportResponse{{
				Organization{Repositories{Nodes: []RepositoriesNode{{IsArchived: false, NameWithOwner: "REPONAME1"}}}},
			}}},
			want: [][]string{{"REPONAME1", "", "false", "false", "false", "false", "", "false", "false", "false"}},
		},
		{
			name: "reportCSVParse branch protection result set",
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
			if got := reportCSVParse(tt.args.ignoreArchived, tt.args.allResults); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportCSVParse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reportCSVLines(t *testing.T) {
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
			name: "reportCSVLines returns no extra rows",
			args: args{parsed: TestEmptyList},
			want: TestEmptyCSVRows,
		},
		{
			name: "reportCSVLines returns some rows",
			args: args{parsed: wantWithBP},
			want: twoCSVRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reportCSVLines(tt.args.parsed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportCSVLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

var errReportCSVFile = errors.New("failed to create report")

func mockReportCSVFile(filePath string, lines [][]string) error {
	return nil
}

func mockReportCSVFileError(filePath string, lines [][]string) error {
	return errReportCSVFile
}

func Test_reportCSVGenerate(t *testing.T) {
	doReportCSVFileWrite = mockReportCSVFile
	defer func() { doReportCSVFileWrite = reportCSVFile }()

	type args struct {
		ignoreArchived bool
		allResults     []ReportResponse
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "reportCSVGenerate success",
			wantErr: false,
		},
		{
			name:    "reportCSVGenerate error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.wantErr {
			doReportCSVFileWrite = mockReportCSVFileError
			defer func() { doReportCSVFileWrite = reportCSVFile }()
		}

		t.Run(tt.name, func(t *testing.T) {
			if err := reportCSVGenerate(tt.args.ignoreArchived, tt.args.allResults); (err != nil) != tt.wantErr {
				t.Errorf("reportCSVGenerate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportCSVFile(t *testing.T) {
	type args struct {
		filePath string
		lines    [][]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "reportCSVFile success",
			args: args{
				filePath: "/tmp/report.csv",
			},
			wantErr: false,
		},
		{
			name: "reportCSVFile error",
			args: args{
				filePath: "/some/dir/doesnt/exist/report.csv",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reportCSVFile(tt.args.filePath, tt.args.lines); (err != nil) != tt.wantErr {
				t.Errorf("reportCSVFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
