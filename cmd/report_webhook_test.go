package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
)

func Test_reportWebhookPostRun(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	originalFilePath := reportWebhookResponse.FilePath

	tests := []struct {
		name          string
		args          args
		filePath      string
		mockJSONError bool
		wantErr       bool
	}{
		{
			name:          "reportWebhookPostRun marshal failure",
			mockJSONError: true,
			wantErr:       true,
		},
		{
			name:     "reportWebhookPostRun fail path failure",
			filePath: "/some/invalid/filepath",
			wantErr:  true,
		},
		{
			name:     "reportWebhookPostRun success",
			filePath: "/tmp/test_report.csv",
			wantErr:  false,
		},
	}

	defer func() {
		jsonMarshal = json.Marshal
		reportWebhookResponse.FilePath = originalFilePath
	}()

	for _, tt := range tests {
		if tt.mockJSONError {
			jsonMarshal = mockJSONMarshalError
		} else {
			jsonMarshal = json.Marshal
		}

		reportWebhookResponse.FilePath = tt.filePath

		t.Run(tt.name, func(t *testing.T) {
			if err := reportWebhookPostRun(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("reportWebhookPostRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportWebhookRun(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		cmd  *cobra.Command
		args []string
	}

	cmdInvalidFlags := &cobra.Command{Use: "report-webhook"}

	cmdAllSetFlags := &cobra.Command{
		Use: "report-webhook",
	}
	cmdAllSetFlags.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdAllSetFlags.Flags().BoolP("ignore-archived", "i", false, "ignore-archived flag")
	cmdAllSetFlags.Flags().StringP(
		"file-path", "f", "report.csv", "File path for report to be created, must be .csv or .json",
	)
	cmdAllSetFlags.Flags().StringP("file-type", "t", "csv", "file type, must be csv or json")
	cmdAllSetFlags.Flags().StringP("start-cursor", "s", "", "The starting cursor for webhook search to start from")
	cmdAllSetFlags.Flags().IntP(
		"timeout", "o", 60, "Timeout for script (in minutes), useful when calling from Lambdas",
	)

	tests := []struct {
		name                 string
		args                 args
		mockHTTPReturnFile   string
		mockHTTPURL          string
		mockHTTPStatusCode   int
		setWebhookResponders bool
		wantErr              bool
	}{
		{
			name: "reportWebhookRun fails on validation",
			args: args{
				cmd: cmdInvalidFlags,
			},
			wantErr: true,
		},
		{
			name: "reportWebhookRun fails on rate limit",
			args: args{
				cmd: cmdAllSetFlags,
			},
			mockHTTPReturnFile: "../testdata/blank.json",
			mockHTTPURL:        "https://api.github.com/rate_limit",
			mockHTTPStatusCode: 401,
			wantErr:            true,
		},
		{
			name: "reportWebhookRun success",
			args: args{
				cmd: cmdAllSetFlags,
			},
			mockHTTPReturnFile:   mockRateLimitResponseFile,
			mockHTTPURL:          "https://api.github.com/rate_limit",
			mockHTTPStatusCode:   200,
			setWebhookResponders: true,
			wantErr:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockHTTPReturnFile != "" {
				mockHTTPResponder("GET", tt.mockHTTPURL, tt.mockHTTPReturnFile, tt.mockHTTPStatusCode)

				if tt.setWebhookResponders {
					mockHTTPResponder(
						"POST",
						"https://api.github.com/graphql",
						"../testdata/mockGraphqlWebhookRepoResponse.json",
						200,
					)
				}
			}

			if err := reportWebhookRun(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("reportWebhookRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportWebhookValidateFlags(t *testing.T) {
	type args struct {
		r   *reportWebhook
		cmd *cobra.Command
	}

	cmdInvalidDryRun := &cobra.Command{Use: "report-webook"}

	cmdInvalidIgnoreArchived := &cobra.Command{Use: "report-webook"}
	cmdInvalidIgnoreArchived.Flags().BoolP("dry-run", "d", false, "dry run flag")

	cmdInvalidFilePath := &cobra.Command{Use: "report-webook"}
	cmdInvalidFilePath.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdInvalidFilePath.Flags().BoolP("ignore-archived", "i", false, "ignore-archived flag")

	cmdInvalidFileType := &cobra.Command{Use: "report-webook"}
	cmdInvalidFileType.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdInvalidFileType.Flags().BoolP("ignore-archived", "i", false, "ignore-archived flag")
	cmdInvalidFileType.Flags().StringP(
		"file-path", "f", "report.csv", "File path for report to be created, must be .csv or .json",
	)

	cmdInvalidStartCursor := &cobra.Command{Use: "report-webook"}
	cmdInvalidStartCursor.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdInvalidStartCursor.Flags().BoolP("ignore-archived", "i", false, "ignore-archived flag")
	cmdInvalidStartCursor.Flags().StringP(
		"file-path", "f", "report.csv", "File path for report to be created, must be .csv or .json",
	)
	cmdInvalidStartCursor.Flags().StringP("file-type", "t", "csv", "file type, must be csv or json")

	cmdInvalidTimeout := &cobra.Command{Use: "report-webook"}
	cmdInvalidTimeout.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdInvalidTimeout.Flags().BoolP("ignore-archived", "i", false, "ignore-archived flag")
	cmdInvalidTimeout.Flags().StringP(
		"file-path", "f", "report.csv", "File path for report to be created, must be .csv or .json",
	)
	cmdInvalidTimeout.Flags().StringP("file-type", "t", "csv", "file type, must be csv or json")
	cmdInvalidTimeout.Flags().StringP("start-cursor", "s", "", "The starting cursor for webhook search to start from")

	cmdInvalid60Timeout := &cobra.Command{Use: "report-webook"}
	cmdInvalid60Timeout.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdInvalid60Timeout.Flags().BoolP("ignore-archived", "i", false, "ignore-archived flag")
	cmdInvalid60Timeout.Flags().StringP(
		"file-path", "f", "report.csv", "File path for report to be created, must be .csv or .json",
	)
	cmdInvalid60Timeout.Flags().StringP("file-type", "t", "csv", "file type, must be csv or json")
	cmdInvalid60Timeout.Flags().StringP("start-cursor", "s", "", "The starting cursor for webhook search to start from")
	cmdInvalid60Timeout.Flags().IntP(
		"timeout", "o", 70, "Timeout for script (in minutes), useful when calling from Lambdas",
	)

	cmdValid := &cobra.Command{Use: "report-webook"}
	cmdValid.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdValid.Flags().BoolP("ignore-archived", "i", false, "ignore-archived flag")
	cmdValid.Flags().StringP(
		"file-path", "f", "report.csv", "File path for report to be created, must be .csv or .json",
	)
	cmdValid.Flags().StringP("file-type", "t", "csv", "file type, must be csv or json")
	cmdValid.Flags().StringP("start-cursor", "s", "", "The starting cursor for webhook search to start from")
	cmdValid.Flags().IntP(
		"timeout", "o", 60, "Timeout for script (in minutes), useful when calling from Lambdas",
	)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "reportWebhookValidateFlags dry run failure",
			args: args{
				cmd: cmdInvalidDryRun,
				r:   &reportWebhook{},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookValidateFlags ignore archived failure",
			args: args{
				cmd: cmdInvalidIgnoreArchived,
				r:   &reportWebhook{},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookValidateFlags file-path failure",
			args: args{
				cmd: cmdInvalidFilePath,
				r:   &reportWebhook{},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookValidateFlags file-type failure",
			args: args{
				cmd: cmdInvalidFileType,
				r:   &reportWebhook{},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookValidateFlags start-cursor failure",
			args: args{
				cmd: cmdInvalidStartCursor,
				r:   &reportWebhook{},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookValidateFlags timeout failure",
			args: args{
				cmd: cmdInvalidTimeout,
				r:   &reportWebhook{},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookValidateFlags invalid timeout failure",
			args: args{
				cmd: cmdInvalid60Timeout,
				r:   &reportWebhook{},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookValidateFlags invalid timeout failure",
			args: args{
				cmd: cmdValid,
				r:   &reportWebhook{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reportWebhookValidateFlags(tt.args.r, tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("reportWebhookValidateFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportWebhookCreate(t *testing.T) {
	type args struct {
		r *reportWebhook
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "reportWebhookCreate fails to return repositories",
			args: args{
				r: &reportWebhook{
					reportWebhookGetter: &mockReportWebhookGetterService{
						failRepoList: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookCreate dry run success",
			args: args{
				r: &reportWebhook{
					dryRun: true,
					reportWebhookGetter: &mockReportWebhookGetterService{
						failRepoList: false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "reportWebhookCreate get webhooks fail",
			args: args{
				r: &reportWebhook{
					reportWebhookGetter: &mockReportWebhookGetterService{
						failWebhook: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookCreate json generate webhook fail",
			args: args{
				r: &reportWebhook{
					reportWebhookGetter: &mockReportWebhookGetterService{},
					reportJSON: &mockReportJSON{
						failgenerate: true,
					},
					fileType: "json",
				},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookCreate json uploader fail",
			args: args{
				r: &reportWebhook{
					reportWebhookGetter: &mockReportWebhookGetterService{},
					reportJSON: &mockReportJSON{
						failupload: true,
					},
					fileType: "json",
				},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookCreate json success",
			args: args{
				r: &reportWebhook{
					reportWebhookGetter: &mockReportWebhookGetterService{},
					reportJSON:          &mockReportJSON{},
					fileType:            "json",
				},
			},
			wantErr: false,
		},
		{
			name: "reportWebhookCreate csv webhook generate fail",
			args: args{
				r: &reportWebhook{
					reportWebhookGetter: &mockReportWebhookGetterService{},
					reportCSV: &mockReportCSV{
						failOpen: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "reportWebhookCreate csv webhook success",
			args: args{
				r: &reportWebhook{
					reportWebhookGetter: &mockReportWebhookGetterService{},
					reportCSV:           &mockReportCSV{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reportWebhookCreate(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("reportWebhookCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportWebhookGetterService_getRepositoryList(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		report *reportWebhook
	}

	tests := []struct {
		name               string
		r                  *reportWebhookGetterService
		args               args
		mockHTTPReturnFile string
		mockHTTPURL        string
		mockHTTPStatusCode int
		want               []repositoryCursorList
		wantErr            bool
	}{
		{
			name: "getRepositoryList with cursor fails graphql call",
			args: args{
				report: &reportWebhook{
					startCursor: "some-cursor",
				},
			},
			mockHTTPReturnFile: "../testdata/mockEmptyResponse.json",
			mockHTTPURL:        "https://api.github.com/graphql",
			mockHTTPStatusCode: 401,
			wantErr:            true,
		},
		{
			name: "getRepositoryList dry run",
			args: args{
				report: &reportWebhook{
					dryRun: true,
				},
			},
			mockHTTPReturnFile: "../testdata/mockGraphqlWebhookRepoResponse.json",
			mockHTTPURL:        "https://api.github.com/graphql",
			mockHTTPStatusCode: 200,
			wantErr:            false,
		},
		{
			name: "getRepositoryList with ignore archived",
			args: args{
				report: &reportWebhook{
					ignoreArchived: true,
				},
			},
			mockHTTPReturnFile: "../testdata/mockGraphqlWebhookRepoWithArchivedResponse.json",
			mockHTTPURL:        "https://api.github.com/graphql",
			mockHTTPStatusCode: 200,
			want: []repositoryCursorList{{
				cursor:       "some-cursor",
				repositories: []string{"repo2"},
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockHTTPReturnFile != "" {
				mockHTTPResponder("POST", tt.mockHTTPURL, tt.mockHTTPReturnFile, tt.mockHTTPStatusCode)
			}

			r := &reportWebhookGetterService{}
			got, err := r.getRepositoryList(tt.args.report)

			if (err != nil) != tt.wantErr {
				t.Errorf("reportWebhookGetterService.getRepositoryList() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportWebhookGetterService.getRepositoryList() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_reportWebhookGetterService_getWebhooks(t *testing.T) {
	originalEndSecs := reportWebhookResponse.EndTimeSecs
	originalConfig := config

	httpmock.Activate()

	defer func() {
		httpmock.DeactivateAndReset()

		config = originalConfig
	}()

	config.Org = "some-org"

	type args struct {
		report       *reportWebhook
		repositories []repositoryCursorList
	}

	tests := []struct {
		name                  string
		r                     *reportWebhookGetterService
		args                  args
		rateLimitResponseFile string
		setEndTimeSecs        int64
		setupWebhookCalls     bool
		want                  map[string][]WebhookResponse
		wantErr               bool
	}{
		{
			name:                  "getWebhooks has reached rate limit",
			rateLimitResponseFile: mockRateLimitEmptyResponseFile,
			args: args{
				repositories: []repositoryCursorList{{cursor: "some-cursor", repositories: []string{"repo1"}}},
			},
			want:    make(map[string][]WebhookResponse, 1),
			wantErr: false,
		},
		{
			name:                  "getWebhooks has timeout elapsed",
			rateLimitResponseFile: mockRateLimitResponseFile,
			args: args{
				repositories: []repositoryCursorList{{cursor: "some-cursor", repositories: []string{"repo1"}}},
			},
			setEndTimeSecs: 0,
			want:           make(map[string][]WebhookResponse, 1),
			wantErr:        false,
		},
		{
			name:                  "getWebhooks success",
			rateLimitResponseFile: mockRateLimitResponseFile,
			setEndTimeSecs:        time.Now().Add(10 * time.Minute).Unix(),
			args: args{
				repositories: []repositoryCursorList{
					{
						cursor:       "some-cursor",
						repositories: []string{"repo1"},
					},
					{
						cursor:       "some-cursor2",
						repositories: []string{"repo2"},
					},
				},
			},
			setupWebhookCalls: true,
			want: map[string][]WebhookResponse{
				"repo1": {
					{
						Config: WebhookResponseConfig{
							URL:         "https://trigger.some_url.com/json",
							InsecureURL: 0,
						},
						Active: true,
						ID:     12345670,
						Events: []string{"push"},
					},
					{
						Config: WebhookResponseConfig{
							URL:         "https://www.some_url.com/sync",
							InsecureURL: 0,
						},
						Active: true,
						ID:     12345671,
						Events: []string{"push"},
					},
					{
						Config: WebhookResponseConfig{
							URL:         "https://www.some_url.com/sync",
							InsecureURL: 0,
						},
						Active: true,
						ID:     12345672,
						Events: []string{"issue_comment", "pull_request", "pull_request_review_comment", "push"},
					},
				},
				"repo2": {
					{
						Config: WebhookResponseConfig{
							URL:         "https://trigger.some_url.com/json",
							InsecureURL: 0,
						},
						Active: true,
						ID:     12345670,
						Events: []string{"push"},
					},
					{
						Config: WebhookResponseConfig{
							URL:         "https://www.some_url.com/sync",
							InsecureURL: 0,
						},
						Active: true,
						ID:     12345671,
						Events: []string{"push"},
					},
					{
						Config: WebhookResponseConfig{
							URL:         "https://www.some_url.com/sync",
							InsecureURL: 0,
						},
						Active: true,
						ID:     12345672,
						Events: []string{"issue_comment", "pull_request", "pull_request_review_comment", "push"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reportWebhookGetterService{}
			reportWebhookResponse.EndTimeSecs = tt.setEndTimeSecs

			// setup rate limit responder
			mockHTTPResponder("GET", "https://api.github.com/rate_limit", tt.rateLimitResponseFile, 200)

			if tt.setupWebhookCalls {
				for _, cursorList := range tt.args.repositories {
					for _, repoName := range cursorList.repositories {
						mockHTTPResponder(
							"GET",
							fmt.Sprintf("https://api.github.com/repos/some-org/%s/hooks", repoName),
							"../testdata/mockRestWebhookResponse.json",
							200,
						)
					}
				}
			}

			got, err := r.getWebhooks(tt.args.report, tt.args.repositories)
			if (err != nil) != tt.wantErr {
				t.Errorf("reportWebhookGetterService.getWebhooks() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportWebhookGetterService.getWebhooks() = %+v, want %+v", got, tt.want)
			}

			reportWebhookResponse.EndTimeSecs = originalEndSecs
		})
	}
}

func Test_setRateLimit(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tests := []struct {
		name                  string
		rateLimitResponseFile string
		wantErr               bool
	}{
		{
			name:                  "setRateLimit error",
			rateLimitResponseFile: mockRestEmptyBodyResponseFile,
			wantErr:               true,
		},
		{
			name:                  "setRateLimit success",
			rateLimitResponseFile: mockRateLimitResponseFile,
			wantErr:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup rate limit responder
			mockHTTPResponder("GET", "https://api.github.com/rate_limit", tt.rateLimitResponseFile, 200)

			if err := setRateLimit(); (err != nil) != tt.wantErr {
				t.Errorf("setRateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_hasReachedRateLimit(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	tests := []struct {
		name                  string
		rateLimitResponseFile string
		want                  bool
	}{
		{
			name:                  "hasReachedRateLimit cannot parse rate limit",
			rateLimitResponseFile: mockRestEmptyBodyResponseFile,
			want:                  true,
		},
		{
			name:                  "hasReachedRateLimit true",
			rateLimitResponseFile: mockRateLimitEmptyResponseFile,
			want:                  true,
		},
		{
			name:                  "hasReachedRateLimit false",
			rateLimitResponseFile: mockRateLimitResponseFile,
			want:                  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup rate limit responder
			mockHTTPResponder("GET", "https://api.github.com/rate_limit", tt.rateLimitResponseFile, 200)

			if got := hasReachedRateLimit(); got != tt.want {
				t.Errorf("hasReachedRateLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}
