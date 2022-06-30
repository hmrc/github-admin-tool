package cmd

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
)

func Test_dependabotRun(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	var (
		mockDryRun      bool
		mockDryRunFalse bool
		mockReposFile   string
		mockRepos2File  string
	)

	mockCmd := &cobra.Command{
		Use: "dependabot",
	}

	mockCmdWithDryRun := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithDryRun.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")

	mockCmdWithDryRunAndRepos := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithDryRunAndRepos.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunAndRepos.Flags().StringVarP(&mockReposFile, "repos", "r", "", "repos file")

	mockCmdWithDryRunOnNoSecurityUpdates := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithDryRunOnNoSecurityUpdates.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunOnNoSecurityUpdates.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"testdata/two_repo_list.txt",
		"repos file",
	)
	mockCmdWithDryRunOnNoSecurityUpdates.Flags().BoolP(
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)

	mockCmdWithDryRunOnNoOptions := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithDryRunOnNoOptions.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunOnNoOptions.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"testdata/two_repo_list.txt",
		"repos file",
	)
	mockCmdWithDryRunOnNoOptions.Flags().BoolP(
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdWithDryRunOnNoOptions.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	mockCmdWithDryRunOn := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithDryRunOn.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunOn.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"testdata/two_repo_list.txt",
		"repos file",
	)
	mockCmdWithDryRunOn.Flags().BoolP(
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdWithDryRunOn.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	if err := mockCmdWithDryRunOn.Flags().Set("alerts", "true"); err != nil {
		t.Errorf("setting alerts flag errors with error = %v", err)
	}

	mockCmdWithDryRunOff := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithDryRunOff.Flags().BoolVarP(&mockDryRunFalse, "dry-run", "d", false, "dry run flag")
	mockCmdWithDryRunOff.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"testdata/two_repo_list.txt",
		"repos file",
	)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "dependabotRun fails dry run",
			args: args{
				cmd: mockCmd,
			},
			wantErr: true,
		},
		{
			name: "dependabotRun fails repo",
			args: args{
				cmd: mockCmdWithDryRun,
			},
			wantErr: true,
		},
		{
			name: "dependabotRun fails repo list path",
			args: args{
				cmd: mockCmdWithDryRunAndRepos,
			},
			wantErr: true,
		},
		{
			name: "dependabotRun fails with no security-updates flag",
			args: args{
				cmd: mockCmdWithDryRunOnNoSecurityUpdates,
			},
			wantErr: true,
		},
		{
			name: "dependabotRun fails on no options passed in",
			args: args{
				cmd: mockCmdWithDryRunOnNoOptions,
			},
			wantErr: true,
		},
		{
			name: "dependabotRun dry run on",
			args: args{
				cmd: mockCmdWithDryRunOn,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dependabotRun(tt.args.cmd, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("dependabotRun() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// nolint // dont care about cyclomatic linting error for test
func Test_dependabotCommand(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		repo *repository
	}

	originalConfig := config

	httpmock.Activate()

	defer func() {
		httpmock.DeactivateAndReset()

		config = originalConfig
	}()

	config.Org = MockOrgName

	var (
		mockDryRun    bool
		mockReposFile string
	)

	mockCmd := &cobra.Command{
		Use: "dependabot",
	}

	mockCmdWithDryRunAndRepos := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithDryRunAndRepos.Flags().BoolVarP(&mockDryRun, "dry-run", "d", true, "dry run flag")
	mockCmdWithDryRunAndRepos.Flags().StringVarP(&mockReposFile, "repos", "r", "", "repos file")
	mockCmdWithDryRunAndRepos.Flags().BoolP(
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdWithDryRunAndRepos.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	if err := mockCmdWithDryRunAndRepos.Flags().Set("alerts", "true"); err != nil {
		t.Errorf("setting alerts flag errors with error = %v", err)
	}

	mockCmdAlertsOn := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdAlertsOn.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdAlertsOn.Flags().StringVarP(&mockReposFile, "repos", "", "r", "repos file")
	mockCmdAlertsOn.Flags().BoolP(
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdAlertsOn.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	if err := mockCmdAlertsOn.Flags().Set("alerts", "true"); err != nil {
		t.Errorf("setting alerts flag errors with error = %v", err)
	}

	mockCmdSecurityUpdatesOn := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdSecurityUpdatesOn.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdSecurityUpdatesOn.Flags().StringVarP(&mockReposFile, "repos", "", "r", "repos file")
	mockCmdSecurityUpdatesOn.Flags().BoolP(
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdSecurityUpdatesOn.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	if err := mockCmdSecurityUpdatesOn.Flags().Set("security-updates", "true"); err != nil {
		t.Errorf("setting security-updates flag errors with error = %v", err)
	}

	if err := mockCmdSecurityUpdatesOn.Flags().Set("alerts", "true"); err != nil {
		t.Errorf("setting alerts flag errors with error = %v", err)
	}

	mockCmdSecurityUpdatesOnNoAlerts := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdSecurityUpdatesOnNoAlerts.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdSecurityUpdatesOnNoAlerts.Flags().StringVarP(&mockReposFile, "repos", "", "r", "repos file")
	mockCmdSecurityUpdatesOnNoAlerts.Flags().BoolP(
		"alerts",
		"a",
		false,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdSecurityUpdatesOnNoAlerts.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	if err := mockCmdSecurityUpdatesOnNoAlerts.Flags().Set("security-updates", "true"); err != nil {
		t.Errorf("setting security-updates flag errors with error = %v", err)
	}

	mockCmdWithFalseAlertFlag := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdWithFalseAlertFlag.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdWithFalseAlertFlag.Flags().StringVarP(&mockReposFile, "repos", "r", "", "repos file")
	mockCmdWithFalseAlertFlag.Flags().BoolP(
		"alerts",
		"a",
		false,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdWithFalseAlertFlag.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	if err := mockCmdWithFalseAlertFlag.Flags().Set("alerts", "true"); err != nil {
		t.Errorf("setting alerts flag errors with error = %v", err)
	}

	mockCmdAlertsOff := &cobra.Command{
		Use: "dependabot",
	}
	mockCmdAlertsOff.Flags().BoolVarP(&mockDryRun, "dry-run", "d", false, "dry run flag")
	mockCmdAlertsOff.Flags().StringVarP(&mockReposFile, "repos", "", "r", "repos file")
	mockCmdAlertsOff.Flags().BoolP(
		"alerts",
		"a",
		false,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdAlertsOff.Flags().BoolP(
		"security-updates",
		"s",
		true,
		"boolean indicating the status of dependabot security updates setting",
	)

	if err := mockCmdAlertsOff.Flags().Set("alerts", "false"); err != nil {
		t.Errorf("setting alerts flag errors with error = %v", err)
	}

	tests := []struct {
		name                 string
		args                 args
		mockHTTPMethod       string
		mockHTTPURL          string
		mockHTTPResponseFile string
		mockHTTPStatusCode   int
		mockUpdateCall       bool
		wantErr              bool
	}{
		{
			name: "dependabotCommand flag check error",
			args: args{
				cmd: mockCmd,
			},
			wantErr: true,
		},
		{
			name: "dependabotCommand no repo file error",
			args: args{
				cmd: mockCmdWithDryRunAndRepos,
				repo: &repository{
					reader: &mockRepositoryReader{
						readFail: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "dependabotCommand is successful with alerts update",
			args: args{
				cmd: mockCmdAlertsOn,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			mockHTTPMethod:       "PUT",
			mockHTTPURL:          "/repos/some-org/some-repo/vulnerability-alerts",
			mockHTTPResponseFile: "testdata/mockRest20xEmptyResponse.json",
			mockHTTPStatusCode:   204,
			wantErr:              false,
		},
		{
			name: "dependabotCommand errors with alerts update",
			args: args{
				cmd: mockCmdAlertsOn,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			mockHTTPMethod:       "PUT",
			mockHTTPURL:          "/repos/some-org/some-repo/vulnerability-alerts",
			mockHTTPResponseFile: "testdata/mockRest404Response.json",
			mockHTTPStatusCode:   404,
			wantErr:              true,
		},
		{
			name: "dependabotCommand is successful with security updates update",
			args: args{
				cmd: mockCmdSecurityUpdatesOn,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			mockHTTPMethod:       "PUT",
			mockHTTPURL:          "/repos/some-org/some-repo/automated-security-fixes",
			mockHTTPResponseFile: "testdata/mockRest20xEmptyResponse.json",
			mockHTTPStatusCode:   204,
			mockUpdateCall:       true,
			wantErr:              false,
		},
		{
			name: "dependabotCommand errors with security updates update",
			args: args{
				cmd: mockCmdSecurityUpdatesOn,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			mockHTTPMethod:       "PUT",
			mockHTTPURL:          "/repos/some-org/some-repo/automated-security-fixes",
			mockHTTPResponseFile: "testdata/mockRest404Response.json",
			mockHTTPStatusCode:   404,
			mockUpdateCall:       true,
			wantErr:              true,
		},
		{
			name: "dependabotCommand errors with security updates update with no alerts",
			args: args{
				cmd: mockCmdSecurityUpdatesOnNoAlerts,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "dependabotCommand is successful with false alert flag",
			args: args{
				cmd: mockCmdWithFalseAlertFlag,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			mockHTTPMethod:       "DELETE",
			mockHTTPURL:          "/repos/some-org/some-repo/automated-security-fixes",
			mockHTTPResponseFile: "testdata/mockRest20xEmptyResponse.json",
			mockHTTPStatusCode:   204,
			mockUpdateCall:       true,
			wantErr:              false,
		},
		{
			name: "dependabotCommand is successful with alerts off",
			args: args{
				cmd: mockCmdAlertsOff,
				repo: &repository{
					reader: &mockRepositoryReader{
						returnValue: []string{
							"some-repo",
						},
					},
				},
			},
			mockHTTPMethod:       "DELETE",
			mockHTTPURL:          "/repos/some-org/some-repo/vulnerability-alerts",
			mockHTTPResponseFile: "testdata/mockRest20xEmptyResponse.json",
			mockHTTPStatusCode:   204,
			wantErr:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockHTTPResponseFile != "" {
				if tt.mockUpdateCall {
					mockHTTPResponder(
						"PUT",
						"/repos/some-org/some-repo/vulnerability-alerts",
						"testdata/mockRest20xEmptyResponse.json",
						204,
					)
				}
				mockHTTPResponder(
					tt.mockHTTPMethod,
					tt.mockHTTPURL,
					tt.mockHTTPResponseFile,
					tt.mockHTTPStatusCode,
				)
			}

			if err := dependabotCommand(tt.args.cmd, tt.args.repo); (err != nil) != tt.wantErr {
				t.Errorf("dependabotCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dependabotToggleAlerts(t *testing.T) {
	type args struct {
		ctx            context.Context
		repositoryName string
		method         string
	}

	originalConfig := config

	httpmock.Activate()

	defer func() {
		httpmock.DeactivateAndReset()

		config = originalConfig
	}()

	config.Org = MockOrgName

	ctx := context.Background()

	tests := []struct {
		name                 string
		args                 args
		mockHTTPMethod       string
		mockHTTPURL          string
		mockHTTPResponseFile string
		mockHTTPStatusCode   int
		wantErr              bool
	}{
		{
			name: "dependabotToggleAlerts errors with delete method",
			args: args{
				ctx:            ctx,
				repositoryName: "some-repo",
				method:         "DELETE",
			},
			mockHTTPMethod:       "DELETE",
			mockHTTPURL:          "/repos/some-org/some-repo/vulnerability-alerts",
			mockHTTPResponseFile: "testdata/mockRest404Response.json",
			mockHTTPStatusCode:   404,
			wantErr:              true,
		},
		{
			name: "dependabotToggleAlerts is successful",
			args: args{
				ctx:            ctx,
				repositoryName: "some-repo",
				method:         "DELETE",
			},
			mockHTTPURL:          "/repos/some-org/some-repo/vulnerability-alerts",
			mockHTTPResponseFile: "testdata/mockRest20xEmptyResponse.json",
			mockHTTPStatusCode:   204,
			wantErr:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPResponder(
				tt.args.method,
				tt.mockHTTPURL,
				tt.mockHTTPResponseFile,
				tt.mockHTTPStatusCode,
			)
			if err := dependabotToggleAlerts(tt.args.ctx, tt.args.repositoryName, tt.args.method); (err != nil) != tt.wantErr {
				t.Errorf("dependabotToggleAlerts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dependabotToggleSecurityUpdates(t *testing.T) {
	type args struct {
		ctx            context.Context
		repositoryName string
		method         string
	}

	originalConfig := config

	httpmock.Activate()

	defer func() {
		httpmock.DeactivateAndReset()

		config = originalConfig
	}()

	config.Org = MockOrgName

	ctx := context.Background()

	tests := []struct {
		name                 string
		args                 args
		mockHTTPURL          string
		mockHTTPResponseFile string
		mockHTTPStatusCode   int
		wantErr              bool
	}{
		{
			name: "dependabotToggleSecurityUpdates errors with delete method",
			args: args{
				ctx:            ctx,
				repositoryName: "some-repo",
				method:         "DELETE",
			},
			mockHTTPURL:          "/repos/some-org/some-repo/automated-security-fixes",
			mockHTTPResponseFile: "testdata/mockRest404Response.json",
			mockHTTPStatusCode:   404,
			wantErr:              true,
		},
		{
			name: "dependabotToggleSecurityUpdates is successful",
			args: args{
				ctx:            ctx,
				repositoryName: "some-repo",
				method:         "DELETE",
			},
			mockHTTPURL:          "/repos/some-org/some-repo/automated-security-fixes",
			mockHTTPResponseFile: "testdata/mockRest20xEmptyResponse.json",
			mockHTTPStatusCode:   204,
			wantErr:              false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPResponder(
				tt.args.method,
				tt.mockHTTPURL,
				tt.mockHTTPResponseFile,
				tt.mockHTTPStatusCode,
			)
			if err := dependabotToggleSecurityUpdates(
				tt.args.ctx,
				tt.args.repositoryName,
				tt.args.method,
			); (err != nil) != tt.wantErr {
				t.Errorf("dependabotToggleSecurityUpdates() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dependabotHTTPMethod(t *testing.T) {
	type args struct {
		enable bool
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "dependabotHTTPMethod return delete",
			want: "DELETE",
		},
		{
			name: "dependabotHTTPMethod return put",
			args: args{
				enable: true,
			},
			want: "PUT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dependabotHTTPMethod(tt.args.enable); got != tt.want {
				t.Errorf("dependabotHTTPMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}
