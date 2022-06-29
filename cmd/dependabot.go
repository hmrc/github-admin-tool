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
	errEmptyFlags                 = errors.New("must set option to update alerts or security-updates or both")
	errFlagsInvalid               = errors.New("alerts must be enabled to configure automated security updates")
	dependabotCmd                 = &cobra.Command{ // nolint // needed for cobra
		Use:   "dependabot",
		Short: "\nEnable and disable dependabot alerts and updates for repos in provided list",
		RunE:  dependabotRun,
	}
)

func dependabotRun(cmd *cobra.Command, args []string) error {
	err := dependabotCommand(
		cmd,
		&repository{
			reader: &repositoryReaderService{},
		},
	)

	return err
}

func dependabotCommand(cmd *cobra.Command, repo *repository) error {
	reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, err := dependabotFlagCheck(cmd)
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
		if isAlertsFlagSet {
			if err := dependabotToggleAlerts(ctx, repositoryName, dependabotHTTPMethod(dependabotAlertsFlag)); err != nil {
				return fmt.Errorf("%w", err)
			}
		}

		if isSecurityUpdatesFlagSet {
			if err := dependabotToggleSecurityUpdates(
				ctx,
				repositoryName,
				dependabotHTTPMethod(dependabotSecurityUpdatesFlag),
			); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	return nil
}

func dependabotToggleAlerts(ctx context.Context, repositoryName, method string) error {
	client := restclient.NewClient(
		fmt.Sprintf("/repos/%s/%s/vulnerability-alerts", config.Org, repositoryName),
		config.Token,
		method,
	)

	var response interface{}

	if err := client.Run(ctx, response); err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Printf(
		"Successful setting to '%s' dependabot alerts for repo %s",
		dependabotStatus(dependabotAlertsFlag),
		repositoryName,
	)

	return nil
}

func dependabotToggleSecurityUpdates(ctx context.Context, repositoryName, method string) error {
	client := restclient.NewClient(
		fmt.Sprintf("/repos/%s/%s/automated-security-fixes", config.Org, repositoryName),
		config.Token,
		method,
	)

	var response interface{}

	if err := client.Run(ctx, response); err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Printf(
		"Successful setting to '%s' dependabot security updates for repo %s",
		dependabotStatus(dependabotSecurityUpdatesFlag),
		repositoryName,
	)

	return nil
}

func dependabotFlagCheck(cmd *cobra.Command) (
	reposFilePath string,
	isAlertsFlagSet,
	isSecurityUpdatesFlagSet,
	dryRun bool,
	err error,
) {
	dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, fmt.Errorf("%w", err)
	}

	reposFilePath, err = cmd.Flags().GetString("repos")
	if err != nil {
		return reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, fmt.Errorf("%w", err)
	}

	dependabotAlertsFlag, err = cmd.Flags().GetBool("alerts")
	if err != nil {
		return reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, fmt.Errorf("%w", err)
	}

	dependabotSecurityUpdatesFlag, err = cmd.Flags().GetBool("security-updates")
	if err != nil {
		return reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, fmt.Errorf("%w", err)
	}

	isAlertsFlagSet = cmd.Flags().Changed("alerts")
	isSecurityUpdatesFlagSet = cmd.Flags().Changed("security-updates")

	if !isAlertsFlagSet && !isSecurityUpdatesFlagSet {
		return reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, errEmptyFlags
	}

	if (isSecurityUpdatesFlagSet) && (!isAlertsFlagSet || !dependabotAlertsFlag) {
		return reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, errFlagsInvalid
	}

	return reposFilePath, isAlertsFlagSet, isSecurityUpdatesFlagSet, dryRun, nil
}

func dependabotStatus(status bool) string {
	if status {
		return "ON"
	}

	return "OFF"
}

func dependabotHTTPMethod(enable bool) string {
	if !enable {
		return http.MethodDelete
	}

	return http.MethodPut
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
