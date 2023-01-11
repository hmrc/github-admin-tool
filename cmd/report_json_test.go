package cmd

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func Test_reportJSONService_uploader(t *testing.T) {
	type args struct {
		filePath   string
		reportJSON []byte
	}

	tests := []struct {
		name    string
		r       *reportJSONService
		args    args
		wantErr bool
	}{
		{
			name: "reportJSONService_uploader success",
			args: args{
				filePath: "/tmp/report.csv",
			},
			wantErr: false,
		},
		{
			name: "reportJSONService_uploader error",
			args: args{
				filePath: "/some/dir/doesnt/exist/report.csv",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reportJSONService{}
			if err := r.uploader(tt.args.filePath, tt.args.reportJSON); (err != nil) != tt.wantErr {
				t.Errorf("reportJSONService.uploader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_reportJSONService_generate(t *testing.T) {
	type args struct {
		ignoreArchived bool
		allResults     []ReportResponse
		teamAccess     map[string]string
	}

	tests := []struct {
		name     string
		r        *reportJSONService
		args     args
		wantFile string
		wantErr  bool
	}{
		{
			name:     "reportJSONService_generate error",
			wantErr:  true,
			wantFile: "testdata/blank.json",
		},
		{
			name: "reportJSONService_generate is success",
			args: args{
				allResults: []ReportResponse{{
					Organization{
						Repositories{
							Nodes: []RepositoriesNode{{
								IsArchived:     false,
								Name:           "REPONAME2",
								NameWithOwner:  "org/REPONAME2",
								HasWikiEnabled: false,
							}},
						},
					},
				}},
			},
			wantFile: "testdata/generate_one_repo.json",
		},
		{
			name: "reportJSONService_generate is success with one archived",
			args: args{
				ignoreArchived: true,
				allResults: []ReportResponse{
					{
						Organization{
							Repositories{
								Nodes: []RepositoriesNode{{
									IsArchived:     false,
									Name:           "REPONAME2",
									NameWithOwner:  "org/REPONAME2",
									HasWikiEnabled: false,
								}},
							},
						},
					},
					{
						Organization{
							Repositories{
								Nodes: []RepositoriesNode{{
									IsArchived:    true,
									Name:          "REPONAME3",
									NameWithOwner: "org/REPONAME3",
								}},
							},
						},
					},
				},
			},
			wantFile: "testdata/generate_one_repo.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reportJSONService{}
			got, err := r.generate(tt.args.ignoreArchived, tt.args.allResults, tt.args.teamAccess)
			if (err != nil) != tt.wantErr {
				t.Errorf("reportJSONService.generate() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			mockReturn, err := ioutil.ReadFile(tt.wantFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			want := string(mockReturn)

			if !reflect.DeepEqual(string(got), want) {
				t.Errorf("reportJSONService.generate() = %v, want %v", string(got), want)
			}
		})
	}
}

func Test_reportJSONService_generateWebhook(t *testing.T) {
	type args struct {
		allResults []Webhooks
	}

	tests := []struct {
		name     string
		r        *reportJSONService
		args     args
		wantFile string
		wantErr  bool
	}{
		{
			name:     "reportJSONService_generateWebhook error",
			wantErr:  true,
			wantFile: "testdata/blank.json",
		},
		{
			name: "reportJSONService_generateWebhook is success",
			args: args{
				allResults: []Webhooks{{
					RepositoryName: "repo1",
					Webhooks: []WebhookResponse{{
						Config: WebhookResponseConfig{
							URL: "some_url", InsecureURL: 0,
						},
						Events: []string{
							"an_event",
						},
					}},
				}},
			},
			wantFile: "testdata/generate_one_webhook_response.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reportJSONService{}
			got, err := r.generateWebhook(tt.args.allResults)
			if (err != nil) != tt.wantErr {
				t.Errorf("reportJSONService.generateWebhook() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			mockReturn, err := ioutil.ReadFile(tt.wantFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			want := string(mockReturn)

			if !reflect.DeepEqual(string(got), want) {
				t.Errorf("reportJSONService.generateWebhook() = %v, want %v", string(got), want)
			}
		})
	}
}
