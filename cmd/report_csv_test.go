package cmd

import (
	"os"
	"reflect"
	"testing"
)

func Test_reportCSVParse(t *testing.T) {
	var (
		emptyAllResults []ReportResponse
		testEmptyList   [][]string
	)

	type args struct {
		ignoreArchived bool
		allResults     []ReportResponse
		teamAccess     map[string]string
	}

	teamAccess := make(map[string]string, 1)
	teamAccess["REPONAME2"] = "ADMIN"

	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "reportCSVParse empty list return empty",
			args: args{ignoreArchived: false, allResults: emptyAllResults},
			want: testEmptyList,
		},
		{
			name: "reportCSVParse archived result set",
			args: args{ignoreArchived: true, allResults: []ReportResponse{{
				Organization{Repositories{Nodes: []RepositoriesNode{{IsArchived: true}}}},
			}}},
			want: testEmptyList,
		},
		{
			name: "reportCSVParse unarchived result set",
			args: args{ignoreArchived: true, allResults: []ReportResponse{{
				Organization{Repositories{Nodes: []RepositoriesNode{{IsArchived: false, NameWithOwner: "REPONAME1"}}}},
			}}},
			want: [][]string{{"REPONAME1", "", "false", "false", "false", "false", "", "false", "false", "false", ""}},
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
				"", "", "false", "false", "false", "false", "", "false", "false", "false", "", "false",
				"false", "false", "false", "false", "false", "false", "false", "0", "false", "false", "SOMEREGEXP",
			}},
		},
		{
			name: "reportCSVParse with team access",
			args: args{
				ignoreArchived: true,
				allResults: []ReportResponse{{
					Organization{
						Repositories{
							Nodes: []RepositoriesNode{{
								IsArchived:    false,
								Name:          "REPONAME2",
								NameWithOwner: "org/REPONAME2",
							}},
						},
					},
				}},
				teamAccess: teamAccess,
			},

			want: [][]string{{"org/REPONAME2", "", "false", "false", "false", "false", "", "false", "false", "false", "ADMIN"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reportCSVParse(
				tt.args.ignoreArchived,
				tt.args.allResults,
				tt.args.teamAccess,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportCSVParse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reportCSVLines(t *testing.T) {
	var testEmptyList [][]string

	wantWithBP := make([][]string, 1)
	wantWithBP[0] = append(
		wantWithBP[0],
		"REPONAME1", "", "false", "false", "false", "false", "", "false", "false",
		"false", "false", "false", "false", "false", "false", "false", "false",
		"false", "0", "false", "false", "SOMEREGEXP",
	)

	twoCSVRows := mockEmptyCSVReportRows
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
			args: args{parsed: testEmptyList},
			want: mockEmptyCSVReportRows,
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

func Test_reportCSVService_opener(t *testing.T) {
	type args struct {
		filePath string
	}

	tests := []struct {
		name    string
		r       *reportCSVService
		args    args
		wantErr bool
	}{
		{
			name: "reportCSVFile opener success",
			args: args{
				filePath: "/tmp/report.csv",
			},
			wantErr: false,
		},
		{
			name: "reportCSVFile opener error",
			args: args{
				filePath: "/some/dir/doesnt/exist/report.csv",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.r.opener(tt.args.filePath); (err != nil) != tt.wantErr {
				t.Errorf("reportCSVFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportCSVService_writer(t *testing.T) {
	type args struct {
		file  *os.File
		lines [][]string
	}

	errFileName := "../testdata/mockFileReadButNotWrite.txt"
	defer os.Chmod(errFileName, 0700) // nolint // only testing so won't check err

	nonReadFile, err := os.Open(errFileName)
	if err != nil {
		t.Error("could not open file")

		return
	}
	os.Chmod(errFileName, 0000) // nolint // only testing so won't check err

	newFile, err := os.Create("/tmp/test_file.txt") // nolint // only testing so ignore
	if err != nil {
		t.Error("could not create file")

		return
	}

	tests := []struct {
		name    string
		r       *reportCSVService
		args    args
		wantErr bool
	}{
		{
			name: "reportCSVFile writer failure",
			args: args{
				file:  nonReadFile,
				lines: [][]string{{"blah de blah"}},
			},
			wantErr: true,
		},
		{
			name: "reportCSVFile writer success",
			args: args{
				file:  newFile,
				lines: [][]string{{"blah de blah"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.writer(tt.args.file, tt.args.lines); (err != nil) != tt.wantErr {
				t.Errorf("reportCSVFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportCSVGenerate(t *testing.T) {
	type args struct {
		ignoreArchived bool
		allResults     []ReportResponse
		teamAccess     map[string]string
	}

	tests := []struct {
		name string
		r    *reportCSVService
		args args
		want [][]string
	}{
		{
			name: "generator returns lines",
			want: mockEmptyCSVReportRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reportCSVGenerate(
				tt.args.ignoreArchived,
				tt.args.allResults,
				tt.args.teamAccess,
			); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportCSVService.generator() = \n%v\n%v\n", got, tt.want)
			}
		})
	}
}

func Test_reportCSVWebhookGenerate(t *testing.T) {
	type args struct {
		webhooks map[string][]WebhookResponse
	}

	tests := []struct {
		name string
		args args
		want [][]string
	}{
		{
			name: "reportCSVWebhookGenerate",
			args: args{
				webhooks: map[string][]WebhookResponse{
					"repo1": {
						WebhookResponse{
							Config: WebhookResponseConfig{
								URL: "some_url", InsecureURL: 0,
							},
							Events: []string{
								"an_event",
							},
						},
					},
				},
			},
			want: [][]string{
				{"Repo Name", "Webhook ID", "Webhook URL", "Is Active", "Insecure URL", "Events"},
				{"repo1", "0", "some_url", "false", "0", "[an_event]"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reportCSVWebhookGenerate(tt.args.webhooks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportCSVWebhookGenerate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reportCSVUpload(t *testing.T) {
	type args struct {
		service  reportCSV
		filePath string
		lines    [][]string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "reportCSVUpload fails on open",
			args: args{
				service: &mockReportCSV{
					failOpen: true,
				},
			},
			wantErr: true,
		},
		{
			name: "reportCSVUpload fails on write",
			args: args{
				service: &mockReportCSV{
					failWrite: true,
				},
			},
			wantErr: true,
		},
		{
			name: "reportCSVUpload success",
			args: args{
				service: &mockReportCSV{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reportCSVUpload(tt.args.service, tt.args.filePath, tt.args.lines); (err != nil) != tt.wantErr {
				t.Errorf("reportCSVUpload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
