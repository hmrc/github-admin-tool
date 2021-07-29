package cmd

import (
	"errors"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
)

func Test_reportGetterService_getReport(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	originalDryRun := dryRun
	defer func() {
		dryRun = originalDryRun
	}()

	var mockEmptyResult []ReportResponse

	tests := []struct {
		name               string
		r                  *reportGetterService
		mockHTTPReturnFile string
		want               []ReportResponse
		dryRunValue        bool
	}{
		{
			name:               "getReport returns empty",
			mockHTTPReturnFile: "testdata/mockEmptyResponse.json",
			want:               nil,
			dryRunValue:        false,
		},
		{
			name:               "getReport dry run true",
			mockHTTPReturnFile: "testdata/mockRepoNodesJsonResponse.json",
			want:               mockEmptyResult,
			dryRunValue:        true,
		},
		{
			name:               "getReport returns one",
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
			dryRunValue: false,
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

			dryRun = tt.dryRunValue

			if got, err := tt.r.getReport(); !reflect.DeepEqual(got, tt.want) {
				if err != nil {
					t.Fatalf("failed to run reportGet %v", err)
				}
				t.Errorf("getReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reportRun(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	var (
		mockDryRun         bool
		mockIgnoreArchived bool
		mockFilePath       string
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

	mockCmdAllFlagsSet := &cobra.Command{
		Use: "report",
	}
	mockCmdAllFlagsSet.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdAllFlagsSet.Flags().BoolVarP(&mockIgnoreArchived, "ignore-archived", "i", true, "ignore flag")
	mockCmdAllFlagsSet.Flags().StringVarP(&mockFilePath, "file-path", "f", "report.csv", "file path flag")

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
			name: "reportRun file-path flag error",
			args: args{
				cmd: mockCmdDryRunOnIgnoreArchived,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: file-path",
		},
		{
			name: "reportRun success",
			args: args{
				cmd: mockCmdAllFlagsSet,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

type mockReportGetter struct {
	fail bool
}

func (m *mockReportGetter) getReport() ([]ReportResponse, error) {
	if m.fail {
		return []ReportResponse{}, errors.New("fail") // nolint // just for testing
	}

	return []ReportResponse{}, nil
}

type mockReportCSV struct {
	fail bool
}

func (m *mockReportCSV) uploader(filePath string, lines [][]string) error {
	if m.fail {
		return errors.New("fail") // nolint // just for testing
	}

	return nil
}

type mockReportAccess struct {
	fail        bool
	returnValue map[string]string
}

func (m *mockReportAccess) getReport() (map[string]string, error) {
	if m.fail {
		return m.returnValue, errors.New("fail") // nolint // just for testing
	}

	return m.returnValue, nil
}

func Test_reportCreate(t *testing.T) {
	type args struct {
		r              *report
		dryRun         bool
		ignoreArchived bool
		filePath       string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "reportCreate failure",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{fail: true},
				},
			},
			wantErr: true,
		},
		{
			name: "reportCreate failure on uploader",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{},
					reportCSV:    &mockReportCSV{fail: true},
					reportAccess: &mockReportAccess{},
				},
			},
			wantErr: true,
		},
		{
			name: "reportCreate success",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{},
					reportCSV:    &mockReportCSV{},
					reportAccess: &mockReportAccess{},
				},
			},
		},
		{
			name: "reportCreate success dry run",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{},
					reportCSV:    &mockReportCSV{},
					reportAccess: &mockReportAccess{},
				},
				dryRun: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reportCreate(
				tt.args.r,
				tt.args.dryRun,
				tt.args.ignoreArchived,
				tt.args.filePath,
			); (err != nil) != tt.wantErr {
				t.Errorf("reportCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
