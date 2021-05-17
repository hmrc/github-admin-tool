package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	repoFile   string
	repos      []string
	signingCmd = &cobra.Command{
		Use:   "signing",
		Short: "Set request signing on to all repos in provided list",
		Run: func(cmd *cobra.Command, args []string) {
			repoFile, _ = cmd.Flags().GetString("repos")
			readList()
			fmt.Printf("repo list is %v", repos)
			//UpdateBranchProtectionRuleInput - requiresCommitSignatures
		},
	}
)

func init() {
	signingCmd.Flags().StringVarP(&repoFile, "repos", "r", "", "repo file")
	signingCmd.MarkFlagRequired("repos")
	rootCmd.AddCommand(signingCmd)
}

func readList() {
	file, err := os.Open(repoFile)
	if err != nil {
		panic(fmt.Errorf("fatal error repo file: %s", err))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repos = append(repos, scanner.Text())
	}
}
