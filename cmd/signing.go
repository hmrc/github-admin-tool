package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	reposFile  string            // nolint // global flag for cobra
	signingCmd = &cobra.Command{ // nolint // needed for cobra
		Use:   "signing",
		Short: "Set request signing on to all repos in provided list",
		Run: func(cmd *cobra.Command, args []string) {
			var (
				err   error
				repos []string
			)

			if reposFile, err = cmd.Flags().GetString("repos"); err != nil {
				log.Fatal(err)
			}

			if repos, err = readList(); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("repo list is %v", repos) // nolint // TODO - REMOVE WHEN FINISHED building
			// UpdateBranchProtectionRuleInput - requiresCommitSignatures
		},
	}
)

// nolint // needed for cobra
func init() {
	signingCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "repo file")
	signingCmd.MarkFlagRequired("repos")
	rootCmd.AddCommand(signingCmd)
}

func readList() ([]string, error) {
	file, err := os.Open(reposFile)

	var repos []string

	if err != nil {
		return repos, fmt.Errorf("fatal error repo file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repos = append(repos, scanner.Text())
	}

	return repos, nil
}
