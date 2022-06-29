package cmd

import (
	"context"
	"errors"
	"fmt"
	"github-admin-tool/restclient"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	dependabotAlertsFlag          bool // nolint // expected global
	dependabotSecurityUpdatesFlag bool // nolint // expected global
	errDependabotFlags            = errors.New("must set option to update alerts or security-updates or both")
	dependabotCmd                 = &cobra.Command{ // nolint // needed for cobra
		Use:   "dependabot",
		Short: "Enable and disable dependabot alerts and updates for repos in provided list",
		RunE:  dependabotRun,
	}
)

const (
	// EmptyFlagCount the number of flags to check empty
	EmptyFlagCount int = 2
)

func dependabotRun(cmd *cobra.Command, args []string) error {
	// command gives a repos file and get alerts flag and security updates flag //
	// we want to check the CSV file input //
	// return a list of repos //
	// check if 1 or more flags are set, otherwise we error //
	// call 2 rest endpoint calls to update dependabot settings //
	err := dependabotCommand(
		cmd,
		&repository{
			reader: &repositoryReaderService{},
		},
	)

	return err
}

func dependabotCommand(cmd *cobra.Command, repo *repository) error {
	reposFilePath, alertsSet, securityUpdatesSet, dryRun, err := dependabotFlagCheck(cmd)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	repositoryList, err := repo.reader.read(reposFilePath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if dryRun {
		log.Printf("This is a dry run, the run would process %d repositories", len(repositoryList))

		return nil
	}

	ctx := context.Background()

	for _, repositoryName := range repositoryList {
		if alertsSet {
			if err := dependabotToggleAlerts(ctx, repositoryName); err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		if securityUpdatesSet {
			if err := dependabotToggleSecurityUpdates(ctx, repositoryName); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	return nil
}

func dependabotToggleAlerts(ctx context.Context, repositoryName string) error {
	method := http.MethodPut
	if !dependabotAlertsFlag {
		method = http.MethodDelete
	}

	log.Printf(
		"Setting to '%s' dependabot alerts for repo %s",
		dependabotStatus(dependabotAlertsFlag),
		repositoryName,
	)

	client := restclient.NewClient(
		fmt.Sprintf("/repos/%s/%s/vulnerability-alerts", config.Org, repositoryName),
		config.Token,
		method,
	)

	var response interface{}

	if err := client.Run(ctx, response); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func dependabotToggleSecurityUpdates(ctx context.Context, repositoryName string) error {
	method := http.MethodPut
	if !dependabotSecurityUpdatesFlag {
		method = http.MethodDelete
	}

	log.Printf(
		"Setting to '%s' dependabot security updates for repo %s",
		dependabotStatus(dependabotSecurityUpdatesFlag),
		repositoryName,
	)

	client := restclient.NewClient(
		fmt.Sprintf("/repos/%s/%s/automated-security-fixes", config.Org, repositoryName),
		config.Token,
		method,
	)

	var response interface{}

	if err := client.Run(ctx, response); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func dependabotFlagCheck(cmd *cobra.Command) (
	reposFilePath string,
	dependabotAlertsSet,
	dependabotSecurityUpdatesSet,
	dryRun bool,
	err error,
) {
	dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return reposFilePath, dependabotAlertsSet, dependabotSecurityUpdatesSet, dryRun, fmt.Errorf("%w", err)
	}

	reposFilePath, err = cmd.Flags().GetString("repos")
	if err != nil {
		return reposFilePath, dependabotAlertsSet, dependabotSecurityUpdatesSet, dryRun, fmt.Errorf("%w", err)
	}

	dependabotAlertsFlag, err = cmd.Flags().GetBool("alerts")
	if err != nil {
		return reposFilePath, dependabotAlertsSet, dependabotSecurityUpdatesSet, dryRun, fmt.Errorf("%w", err)
	}

	dependabotSecurityUpdatesFlag, err = cmd.Flags().GetBool("security-updates")
	if err != nil {
		return reposFilePath, dependabotAlertsSet, dependabotSecurityUpdatesSet, dryRun, fmt.Errorf("%w", err)
	}

	var missingFlags int

	if dependabotAlertsSet = cmd.Flags().Changed("alerts"); !dependabotAlertsSet {
		missingFlags++
	}

	if dependabotSecurityUpdatesSet = cmd.Flags().Changed("security-updates"); !dependabotSecurityUpdatesSet {
		missingFlags++
	}

	if missingFlags == EmptyFlagCount {
		return reposFilePath, dependabotAlertsSet, dependabotSecurityUpdatesSet, dryRun, errDependabotFlags
	}

	return reposFilePath, dependabotAlertsSet, dependabotSecurityUpdatesSet, dryRun, nil
}

func dependabotStatus(status bool) string {
	if status {
		return "ON"
	}

	return "OFF"
}

// nolint // needed for cobra
func init() {
	dependabotCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "path to file containing repositories (file should contain repos on new line without org/ prefix)")
	dependabotCmd.Flags().BoolVarP(&dependabotAlertsFlag, "alerts", "a", true, "boolean indicating the status of dependabot alerts setting")
	dependabotCmd.Flags().BoolVarP(&dependabotSecurityUpdatesFlag, "security-updates", "s", true, "boolean indicating the status of dependabot security updates setting")
	dependabotCmd.MarkFlagRequired("repos")
	dependabotCmd.Flags().SortFlags = true
	rootCmd.AddCommand(dependabotCmd)
}
