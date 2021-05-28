package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	reposFile  string
	repos      []string
	signingCmd = &cobra.Command{
		Use:   "signing",
		Short: "Set request signing on to all repos in provided list",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			reposFile, err = cmd.Flags().GetString("repos")
			if err != nil {
				log.Fatal(err)
			}
			readList()
			fmt.Printf("repo list is %v", repos)
			// UpdateBranchProtectionRuleInput - requiresCommitSignatures
		},
	}
)

func init() {
	signingCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "repo file")
	signingCmd.MarkFlagRequired("repos")
	rootCmd.AddCommand(signingCmd)
}

func readList() {
	file, err := os.Open(reposFile)
	if err != nil {
		panic(fmt.Errorf("fatal error repo file: %w", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repos = append(repos, scanner.Text())
	}
}
