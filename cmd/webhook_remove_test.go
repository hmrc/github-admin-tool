package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
)

func Test_removeWebhook(t *testing.T) {
	originalConfig := config

	httpmock.Activate()

	defer func() {
		httpmock.DeactivateAndReset()

		config = originalConfig
	}()

	config.Org = MockOrgName

	type args struct {
		ctx            context.Context
		webhookID      int
		repositoryName string
	}

	ctx := context.Background()

	tests := []struct {
		name               string
		args               args
		mockHTTPStatusCode int
		wantErr            bool
	}{
		{
			name: "removeWebhook failure",
			args: args{
				ctx:            ctx,
				webhookID:      12456789,
				repositoryName: "some-repo-name",
			},
			mockHTTPStatusCode: 404,
			wantErr:            true,
		},
		{
			name: "removeWebhook success",
			args: args{
				ctx:            ctx,
				webhookID:      12456789,
				repositoryName: "some-repo-name",
			},
			mockHTTPStatusCode: 204,
			wantErr:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPResponder(
				"DELETE",
				fmt.Sprintf("/repos/some-org/%s/hooks/%d", tt.args.repositoryName, tt.args.webhookID),
				"testdata/mockEmptyResponse.json",
				tt.mockHTTPStatusCode,
			)

			if err := removeWebhook(tt.args.ctx, tt.args.webhookID, tt.args.repositoryName); (err != nil) != tt.wantErr {
				t.Errorf("removeWebhook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getWebhookID(t *testing.T) {
	originalConfig := config

	httpmock.Activate()

	defer func() {
		httpmock.DeactivateAndReset()

		config = originalConfig
	}()

	config.Org = MockOrgName

	ctx := context.Background()

	type args struct {
		ctx            context.Context
		webhookURL     string
		repositoryName string
	}

	tests := []struct {
		name                 string
		args                 args
		mockHTTPResponseFile string
		mockHTTPStatusCode   int
		wantWebhookID        int
	}{
		{
			name: "getWebhookID not found",
			args: args{
				ctx:            ctx,
				webhookURL:     "https://some-webhook-host",
				repositoryName: "some-repo-name",
			},
			mockHTTPResponseFile: "testdata/blank.json",
			mockHTTPStatusCode:   404,
		},
		{
			name: "getWebhookID found",
			args: args{
				ctx:            ctx,
				webhookURL:     "https://some-external-webhook.org",
				repositoryName: "some-repo-name2",
			},
			mockHTTPResponseFile: "testdata/mockGetWebhooksResponse.json",
			mockHTTPStatusCode:   200,
			wantWebhookID:        789,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPResponder(
				"GET",
				fmt.Sprintf("/repos/some-org/%s/hooks", tt.args.repositoryName),
				tt.mockHTTPResponseFile,
				tt.mockHTTPStatusCode,
			)
			if gotWebhookID := getWebhookID(
				tt.args.ctx,
				tt.args.webhookURL,
				tt.args.repositoryName,
			); gotWebhookID != tt.wantWebhookID {
				t.Errorf("getWebhookID() = %v, want %v", gotWebhookID, tt.wantWebhookID)
			}
		})
	}
}

func Test_removeWebhookFlagCheck(t *testing.T) {
	type args struct {
		cmd *cobra.Command
	}

	cmdInvalidFlags := &cobra.Command{Use: "webhook-remove"}

	cmdNoWebhookURLFlags := &cobra.Command{Use: "webhook-remove"}
	cmdNoWebhookURLFlags.Flags().BoolP("dry-run", "d", false, "dry run flag")

	cmdInvalidWebhookURLFlags := &cobra.Command{Use: "webhook-remove"}
	cmdInvalidWebhookURLFlags.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdInvalidWebhookURLFlags.Flags().StringP("url", "", "http//invalid-host", "url flag")

	cmdNoReposFlags := &cobra.Command{Use: "webhook-remove"}
	cmdNoReposFlags.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdNoReposFlags.Flags().StringP("url", "", "https://valid-host", "url flag")

	cmdValidFlags := &cobra.Command{Use: "webhook-remove"}
	cmdValidFlags.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdValidFlags.Flags().StringP("url", "", "https://valid-host", "url flag")
	cmdValidFlags.Flags().StringP("repos", "", "filepath", "repos flag")

	tests := []struct {
		name              string
		args              args
		wantWebhookURL    string
		wantReposFilePath string
		wantDryRun        bool
		wantErr           bool
	}{
		{
			name: "removeWebhookFlagCheck fails no dry run",
			args: args{
				cmd: cmdInvalidFlags,
			},
			wantErr: true,
		},
		{
			name: "removeWebhookFlagCheck fails no url",
			args: args{
				cmd: cmdNoWebhookURLFlags,
			},
			wantErr: true,
		},
		{
			name: "removeWebhookFlagCheck fails invalid url",
			args: args{
				cmd: cmdInvalidWebhookURLFlags,
			},
			wantWebhookURL: "http//invalid-host",
			wantErr:        true,
		},
		{
			name: "removeWebhookFlagCheck fails no repos",
			args: args{
				cmd: cmdNoReposFlags,
			},
			wantWebhookURL: "https://valid-host",
			wantErr:        true,
		},
		{
			name: "removeWebhookFlagCheck fails no repos",
			args: args{
				cmd: cmdValidFlags,
			},
			wantWebhookURL:    "https://valid-host",
			wantDryRun:        false,
			wantReposFilePath: "filepath",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWebhookURL, gotReposFilePath, gotDryRun, err := removeWebhookFlagCheck(tt.args.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("removeWebhookFlagCheck() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if gotWebhookURL != tt.wantWebhookURL {
				t.Errorf("removeWebhookFlagCheck() gotWebhookURL = %v, want %v", gotWebhookURL, tt.wantWebhookURL)
			}
			if gotReposFilePath != tt.wantReposFilePath {
				t.Errorf("removeWebhookFlagCheck() gotReposFilePath = %v, want %v", gotReposFilePath, tt.wantReposFilePath)
			}
			if gotDryRun != tt.wantDryRun {
				t.Errorf("removeWebhookFlagCheck() gotDryRun = %v, want %v", gotDryRun, tt.wantDryRun)
			}
		})
	}
}

func Test_removeWebhookCommand(t *testing.T) {
	originalConfig := config

	httpmock.Activate()

	defer func() {
		httpmock.DeactivateAndReset()

		config = originalConfig
	}()

	config.Org = MockOrgName

	type args struct {
		cmd  *cobra.Command
		repo *repository
	}

	cmdInvalidFlags := &cobra.Command{Use: "webhook-remove"}

	cmdDryRunOnFlags := &cobra.Command{Use: "webhook-remove"}
	cmdDryRunOnFlags.Flags().BoolP("dry-run", "d", true, "dry run flag")
	cmdDryRunOnFlags.Flags().StringP("url", "", "https://some-external-webhook.com", "url flag")
	cmdDryRunOnFlags.Flags().StringP("repos", "", "filepath", "repos flag")

	cmdDryRunOffFlags := &cobra.Command{Use: "webhook-remove"}
	cmdDryRunOffFlags.Flags().BoolP("dry-run", "d", false, "dry run flag")
	cmdDryRunOffFlags.Flags().StringP("url", "", "https://some-external-webhook.com", "url flag")
	cmdDryRunOffFlags.Flags().StringP("repos", "", "filepath", "repos flag")

	tests := []struct {
		name         string
		args         args
		mockHTTPFunc func()
		wantErr      bool
	}{
		{
			name: "removeWebhookCommand flag check failure",
			args: args{
				cmd: cmdInvalidFlags,
			},
			mockHTTPFunc: func() {},
			wantErr:      true,
		},
		{
			name: "removeWebhookCommand repo read failure",
			args: args{
				cmd: cmdDryRunOnFlags,
				repo: &repository{
					reader: &mockRepositoryReader{
						readFail: true,
					},
				},
			},
			mockHTTPFunc: func() {},
			wantErr:      true,
		},
		{
			name: "removeWebhookCommand dry run success",
			args: args{
				cmd: cmdDryRunOnFlags,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo-name",
						},
					},
				},
			},
			mockHTTPFunc: func() {},
			wantErr:      false,
		},
		{
			name: "removeWebhookCommand remove webhook error",
			args: args{
				cmd: cmdDryRunOffFlags,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			mockHTTPFunc: func() {
				mockHTTPResponder(
					"GET",
					"/repos/some-org/some-repo/hooks",
					"testdata/mockGetWebhooksResponse.json",
					200,
				)
				mockHTTPResponder(
					"DELETE",
					"/repos/some-org/some-repo/hooks/123",
					"testdata/mockDeleteWebhookResponse404.json",
					404,
				)
			},
			wantErr: true,
		},
		{
			name: "removeWebhookCommand success with empty repo",
			args: args{
				cmd: cmdDryRunOffFlags,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{},
					},
				},
			},
			mockHTTPFunc: func() {},
			wantErr:      false,
		},
		{
			name: "removeWebhookCommand success with remove",
			args: args{
				cmd: cmdDryRunOffFlags,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{"some-repo-name"},
					},
				},
			},
			mockHTTPFunc: func() {
				mockHTTPResponder(
					"GET",
					"/repos/some-org/some-repo/hooks",
					"testdata/mockGetWebhooksResponse.json",
					200,
				)
				mockHTTPResponder(
					"DELETE",
					"/repos/some-org/some-repo/hooks/123",
					"testdata/blank.json",
					200,
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockHTTPFunc()
			if err := removeWebhookCommand(tt.args.cmd, tt.args.repo); (err != nil) != tt.wantErr {
				t.Errorf("removeWebhookCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webhookRemoveRun(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	cmdDryRunOnFlags := &cobra.Command{Use: "webhook-remove"}
	cmdDryRunOnFlags.Flags().BoolP("dry-run", "d", true, "dry run flag")
	cmdDryRunOnFlags.Flags().StringP("url", "", "https://some-external-webhook.com", "url flag")
	cmdDryRunOnFlags.Flags().StringP("repos", "", "testdata/one_repo_list.txt", "repos flag")

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "webhookRemoveRun success",
			args: args{
				cmd: cmdDryRunOnFlags,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := webhookRemoveRun(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("webhookRemoveRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
