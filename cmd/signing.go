package cmd

import (
	"fmt"
	"github-admin-tool/graphqlclient"
	"log"
	"os"

	"github.com/spf13/cobra"
)

const maxRepositories = 100

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

		repoMap, err := readRepoList(reposFilePath)
		if err != nil {
			log.Fatal(err)
		}

		numberOfRepos := len(repoMap)
		if numberOfRepos < 1 || numberOfRepos > maxRepositories {
			log.Fatal("Number of repos passed in must be more than 1 and less than 100")
		}

		queryString := generateRepoQuery(repoMap)
		client := graphqlclient.NewClient("https://api.github.com/graphql")
		repoSearchResult, err := repoRequest(queryString, client)
		if err != nil {
			log.Fatal(err)
		}

		if dryRun {
			log.Printf("This is a dry run, the run would process %d repositories", numberOfRepos)
			os.Exit(0)
		}

		updated, created, info, problems := applySigning(repoSearchResult, client)

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

func applySigning(repoSearchResult map[string]RepositoriesNodeList, client *graphqlclient.Client) (
	modified,
	created,
	info,
	problems []string,
) {
	var err error

OUTER:

	for _, repository := range repoSearchResult { // nolint
		if repository.DefaultBranchRef.Name == "" {
			info = append(info, fmt.Sprintf("No default branch for %v", repository.NameWithOwner))

			continue OUTER
		}

		signingArgs := setSigningArgs()

		// Check all nodes for default branch protection rule
		for _, branchProtection := range repository.BranchProtectionRules.Nodes {
			if repository.DefaultBranchRef.Name != branchProtection.Pattern {
				continue
			}

			// If default branch has already got signing turned on, no need to update
			if branchProtection.RequiresCommitSignatures {
				info = append(info, fmt.Sprintf("Signing already turned on for %v", repository.NameWithOwner))

				continue OUTER
			}

			if err = signingUpdate(branchProtection.ID, signingArgs, client); err != nil {
				problems = append(problems, err.Error())

				continue OUTER
			}
			modified = append(modified, repository.NameWithOwner)

			continue OUTER
		}

		if err = signingCreate(repository.ID, repository.DefaultBranchRef.Name, signingArgs, client); err != nil {
			problems = append(problems, err.Error())

			continue OUTER
		}

		created = append(created, repository.NameWithOwner)
	}

	return modified, created, info, problems
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
