package cmd

import (
	"github-admin-tool/graphqlclient"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var signingCmd = &cobra.Command{ // nolint // needed for cobra
	Use:   "signing",
	Short: "Set request signing on to all repos in provided list",
	Run: func(cmd *cobra.Command, args []string) {
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			log.Fatal(err)
		}

		reposFilePath, err := cmd.Flags().GetString("repos")
		if err != nil {
			log.Fatal(err)
		}

		repoMap, err := repoList(reposFilePath)
		if err != nil {
			log.Fatal(err)
		}

		numberOfRepos := len(repoMap)
		if numberOfRepos < 1 || numberOfRepos > maxRepositories {
			log.Fatal("Number of repos passed in must be more than 1 and less than 100")
		}

		queryString := repoQuery(repoMap)
		client := graphqlclient.NewClient("https://api.github.com/graphql")
		repoSearchResult, err := repoRequest(queryString, client)
		if err != nil {
			log.Fatal(err)
		}

		if dryRun {
			log.Printf("This is a dry run, the run would process %d repositories", numberOfRepos)
			os.Exit(0)
		}

		signingArgs := setSigningArgs()

		updated, created, info, problems := branchProtectionApply(repoSearchResult, "Signing", signingArgs, client)

		for key, repo := range updated {
			log.Printf("Modified (%d): %v", key, repo)
		}

		for key, repo := range created {
			log.Printf("Created (%d): %v", key, repo)
		}

		for key, err := range problems {
			log.Printf("Error (%d): %v", key, err)
		}

		for key, i := range info {
			log.Printf("Info (%d): %v", key, i)
		}
	},
}

// nolint // needed for cobra
func init() {
	signingCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "file containing repositories on new line without org/ prefix. Max 100 repos")
	signingCmd.MarkFlagRequired("repos")
	rootCmd.AddCommand(signingCmd)
}

func setSigningArgs() (branchProtectionArgs []BranchProtectionArgs) {
	branchProtectionArgs = append(
		branchProtectionArgs,
		BranchProtectionArgs{
			Name:     "requiresCommitSignatures",
			DataType: "Boolean",
			Value:    true,
		})

	return branchProtectionArgs
}
