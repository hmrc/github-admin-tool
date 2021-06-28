package cmd

import (
	"errors"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
)

func Test_reportGet(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tests := []struct {
		name               string
		mockHTTPReturnFile string
		want               []ReportResponse
	}{
		{
			name:               "reportGet returns empty",
			mockHTTPReturnFile: "testdata/mockEmptyResponse.json",
			want:               nil,
		},
		{
			name:               "reportGet returns one",
			mockHTTPReturnFile: "testdata/mockRepoNodesJsonResponse.json",
			want: []ReportResponse{{Organization{Repositories{
				TotalCount: 1,
				Nodes: []RepositoriesNode{{
					Name:               "repo-name",
					NameWithOwner:      "org-name/repo-name",
					DefaultBranchRef:   DefaultBranchRef{"main"},
					SquashMergeAllowed: true,
				}},
			}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPReturn, err := ioutil.ReadFile(tt.mockHTTPReturnFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			httpmock.RegisterResponder(
				"POST",
				"https://api.github.com/graphql",
				httpmock.NewStringResponder(200, string(mockHTTPReturn)),
			)

			if got, err := reportGet(); !reflect.DeepEqual(got, tt.want) {
				if err != nil {
					t.Fatalf("failed to run reportGet %v", err)
				}
				t.Errorf("reportGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	errMockReportRequest     = errors.New("report failure")
	errMockReportCSVGenerate = errors.New("report csv generate failure")
)

func mockDoReportGet() (results []ReportResponse, err error) {
	return results, nil
}

func mockDoReportGetError() (results []ReportResponse, err error) {
	return results, errMockReportRequest
}

func mockDoReportCSVGenerate(ignoreArchived bool, allResults []ReportResponse) error {
	return nil
}

func mockDoReportCSVGenerateError(ignoreArchived bool, allResults []ReportResponse) error {
	return errMockReportCSVGenerate
}

func Test_reportRun(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	var (
		mockDryRun         bool
		mockIgnoreArchived bool
	)

	mockCmd := &cobra.Command{
		Use: "report",
	}

	mockCmdDryRunOn := &cobra.Command{
		Use: "report",
	}
	mockCmdDryRunOn.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")

	mockCmdDryRunOnIgnoreArchived := &cobra.Command{
		Use: "report",
	}
	mockCmdDryRunOnIgnoreArchived.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdDryRunOnIgnoreArchived.Flags().BoolVarP(&mockIgnoreArchived, "ignore-archived", "i", true, "ignore flag")

	mockCmdDryRunFalse := &cobra.Command{
		Use: "report",
	}
	mockCmdDryRunFalse.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdDryRunFalse.Flags().BoolVarP(&mockIgnoreArchived, "ignore-archived", "i", true, "ignore flag")

	tests := []struct {
		name                         string
		args                         args
		wantErr                      bool
		wantErrMsg                   string
		mockRequestErrorFunction     bool
		mockCSVGenerateErrorFunction bool
	}{
		{
			name: "reportRun dry run flag error",
			args: args{
				cmd: mockCmd,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: dry-run",
		},
		{
			name: "reportRun ignore-archived flag error",
			args: args{
				cmd: mockCmdDryRunOn,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: ignore-archived",
		},
		{
			name: "reportRun report request failure",
			args: args{
				cmd: mockCmdDryRunOnIgnoreArchived,
			},
			wantErr:                  true,
			wantErrMsg:               "report failure",
			mockRequestErrorFunction: true,
		},
		{
			name: "reportRun generate csv error",
			args: args{
				cmd: mockCmdDryRunFalse,
			},
			wantErr:                      true,
			wantErrMsg:                   "report csv generate failure",
			mockCSVGenerateErrorFunction: true,
		},
		{
			name: "reportRun success",
			args: args{
				cmd: mockCmdDryRunFalse,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doReportGet = mockDoReportGet
			doReportCSVGenerate = mockDoReportCSVGenerate
			if tt.mockRequestErrorFunction {
				doReportGet = mockDoReportGetError
			}
			if tt.mockCSVGenerateErrorFunction {
				doReportCSVGenerate = mockDoReportCSVGenerateError
			}
			defer func() {
				doReportGet = reportGet
				doReportCSVGenerate = reportCSVGenerate
			}()

			err := reportRun(tt.args.cmd, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("reportRun() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("reportRun() error = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}
