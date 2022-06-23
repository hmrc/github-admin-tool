package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	dependabotAlerts          bool // nolint // expected global
	dependabotSecurityUpdates bool // nolint // expected global
	errDependabotOptions      = errors.New("must set option to update alerts or security-updates or both")
	dependabotCmd             = &cobra.Command{ // nolint // needed for cobra
		Use:   "dependabot",
		Short: "Enable and disable dependabot alerts and updates for repos in provided list",
		RunE:  dependabotRun,
	}
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
	// reposFilePath, dependabotAlerts, dependabotSecurityUpdates, dryRun, err := dependabotFlagCheck(cmd)
	reposFilePath, _, _, dryRun, err := dependabotFlagCheck(cmd)
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

	return nil
}

func dependabotFlagCheck(cmd *cobra.Command) (
	reposFilePath string,
	dependabotAlerts,
	dependabotSecurityUpdates,
	dryRun bool,
	err error,
) {
	dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return reposFilePath, dependabotAlerts, dependabotSecurityUpdates, dryRun, fmt.Errorf("%w", err)
	}

	reposFilePath, err = cmd.Flags().GetString("repos")
	if err != nil {
		return reposFilePath, dependabotAlerts, dependabotSecurityUpdates, dryRun, fmt.Errorf("%w", err)
	}

	dependabotAlerts, err = cmd.Flags().GetBool("alerts")
	if err != nil {
		return reposFilePath, dependabotAlerts, dependabotSecurityUpdates, dryRun, fmt.Errorf("%w", err)
	}

	dependabotSecurityUpdates, err = cmd.Flags().GetBool("security-updates")
	if err != nil {
		return reposFilePath, dependabotAlerts, dependabotSecurityUpdates, dryRun, fmt.Errorf("%w", err)
	}

	if !cmd.Flags().Changed("alerts") && !cmd.Flags().Changed("security-updates") {
		return reposFilePath, dependabotAlerts, dependabotSecurityUpdates, dryRun, errDependabotOptions
	}

	return reposFilePath, dependabotAlerts, dependabotSecurityUpdates, dryRun, nil
}

// nolint // needed for cobra
func init() {
	dependabotCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "path to file containing repositories (file should contain repos on new line without org/ prefix)")
	dependabotCmd.Flags().BoolVarP(&dependabotAlerts, "alerts", "a", true, "boolean indicating the status of dependabot alerts setting")
	dependabotCmd.Flags().BoolVarP(&dependabotSecurityUpdates, "security-updates", "s", true, "boolean indicating the status of dependabot security updates setting")
	dependabotCmd.MarkFlagRequired("repos")
	dependabotCmd.Flags().SortFlags = true
	rootCmd.AddCommand(dependabotCmd)
}
