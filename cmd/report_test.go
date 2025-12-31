package cmd

import (
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockHTTPResponder("POST", "https://api.github.com/graphql", test.mockHTTPReturnFile, 200)

			dryRun = test.dryRunValue

			if got, err := test.r.getReport(); !reflect.DeepEqual(got, test.want) {
				if err != nil {
					t.Fatalf("failed to run reportGet %v", err)
				}
				t.Errorf("getReport() = %v, want %v", got, test.want)
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
		mockFileType       string
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

	mockCmdFileTypeMissing := &cobra.Command{
		Use: "report",
	}
	mockCmdFileTypeMissing.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdFileTypeMissing.Flags().BoolVarP(&mockIgnoreArchived, "ignore-archived", "i", true, "ignore flag")
	mockCmdFileTypeMissing.Flags().StringVarP(&mockFilePath, "file-path", "f", "report.csv", "file path flag")

	mockCmdAllFlagsSet := &cobra.Command{
		Use: "report",
	}
	mockCmdAllFlagsSet.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdAllFlagsSet.Flags().BoolVarP(&mockIgnoreArchived, "ignore-archived", "i", true, "ignore flag")
	mockCmdAllFlagsSet.Flags().StringVarP(&mockFilePath, "file-path", "f", "report.csv", "file path flag")
	mockCmdAllFlagsSet.Flags().StringVarP(&mockFileType, "file-type", "t", "csv", "file type flag")

	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
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
			name: "reportRun file-type flag error",
			args: args{
				cmd: mockCmdFileTypeMissing,
			},
			wantErr:    true,
			wantErrMsg: "flag accessed but not defined: file-type",
		},
		{
			name: "reportRun success",
			args: args{
				cmd: mockCmdAllFlagsSet,
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := reportRun(test.args.cmd, test.args.args)
			if (err != nil) != test.wantErr {
				t.Errorf("reportRun() error = %v, wantErr %v", err, test.wantErr)
			}
			if err != nil && test.wantErr && err.Error() != test.wantErrMsg {
				t.Errorf("reportRun() error = %v, wantErrMsg %v", err.Error(), test.wantErrMsg)
			}
		})
	}
}

func Test_reportCreate(t *testing.T) {
	type args struct {
		r              *report
		dryRun         bool
		ignoreArchived bool
		filePath       string
		fileType       string
	}

	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantErrMsg string
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
					reportCSV:    &mockReportCSV{failOpen: true},
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
			name: "reportCreate upload failure with JSON",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{},
					reportCSV:    &mockReportCSV{},
					reportJSON:   &mockReportJSON{failupload: true},
					reportAccess: &mockReportAccess{},
				},
				fileType: "json",
			},
			wantErr:    true,
			wantErrMsg: "upload json failed: fail",
		},
		{
			name: "reportCreate generate failure with JSON",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{},
					reportCSV:    &mockReportCSV{},
					reportJSON:   &mockReportJSON{failgenerate: true},
					reportAccess: &mockReportAccess{},
				},
				fileType: "json",
			},
			wantErr:    true,
			wantErrMsg: "generate json failed: fail",
		},
		{
			name: "reportCreate success with JSON",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{},
					reportCSV:    &mockReportCSV{},
					reportJSON:   &mockReportJSON{},
					reportAccess: &mockReportAccess{},
				},
				fileType: "json",
			},
		},
		{
			name: "reportCreate failure on access report",
			args: args{
				r: &report{
					reportGetter: &mockReportGetter{},
					reportCSV:    &mockReportCSV{},
					reportJSON:   &mockReportJSON{},
					reportAccess: &mockReportAccess{fail: true},
				},
			},
			wantErr:    true,
			wantErrMsg: "access fail",
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := reportCreate(
				test.args.r,
				test.args.dryRun,
				test.args.ignoreArchived,
				test.args.filePath,
				test.args.fileType,
			)
			if (err != nil) != test.wantErr {
				t.Errorf("reportCreate() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.wantErrMsg != "" && test.wantErrMsg != err.Error() {
				t.Errorf("reportCreate() error = %v, wantErrMsg %v", err.Error(), test.wantErrMsg)
			}
		})
	}
}
