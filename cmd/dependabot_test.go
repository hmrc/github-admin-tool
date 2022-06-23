package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func Test_dependabotRun(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}

	mockCmd := &cobra.Command{
		Use: "dependabot",
	}

	mockCmdWithDryRun := &cobra.Command{
		Use: "dependabot",
	}

	var (
		mockDependabotAlerts           bool
		mockDdependabotSecurityUpdates bool
		mockDryRun                     bool
		mockDryRunFalse                bool
		mockReposFile                  string
		mockRepos2File                 string
	)

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
	mockCmdWithDryRunOnNoSecurityUpdates.Flags().BoolVarP(
		&mockDependabotAlerts,
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
	mockCmdWithDryRunOnNoOptions.Flags().BoolVarP(
		&mockDependabotAlerts,
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdWithDryRunOnNoOptions.Flags().BoolVarP(
		&mockDependabotAlerts,
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
	mockCmdWithDryRunOn.Flags().BoolVarP(
		&mockDependabotAlerts,
		"alerts",
		"a",
		true,
		"boolean indicating the status of dependabot alerts setting",
	)
	mockCmdWithDryRunOn.Flags().BoolVarP(
		&mockDdependabotSecurityUpdates,
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
