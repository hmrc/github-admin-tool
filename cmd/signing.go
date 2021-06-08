package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github-admin-tool/graphqlclient"

	"github.com/spf13/cobra"
)

const maxRepositories = 100

var (
	reposFile  string            // nolint // global flag for cobra
	signingCmd = &cobra.Command{ // nolint // needed for cobra
		Use:   "signing",
		Short: "Set request signing on to all repos in provided list",
		Run: func(cmd *cobra.Command, args []string) {
			reposFilePath, err := cmd.Flags().GetString("repos")
			if err != nil {
				log.Fatal(err)
			}

			repoMap, err := readList(reposFilePath)
			if err != nil {
				log.Fatal(err)
			}

			numberOfRepos := len(repoMap)
			if numberOfRepos < 1 || numberOfRepos > maxRepositories {
				log.Fatal("Number of repos passed in must be more than 1 and less than 100")
			}

			queryString := generateQuery(repoMap)
			client := graphqlclient.NewClient("https://api.github.com/graphql")
			repoSearchResult, err := repoRequest(queryString, client)
			if err != nil {
				log.Fatal(err)
			}

			updated, info, errors := applySigning(repoSearchResult, client)

			for key, repo := range updated {
				log.Printf("Modified Repo (%d): %v", key, repo)
			}

			for key, err := range errors {
				log.Printf("Error (%d): %v", key, err)
			}

			for key, i := range info {
				log.Printf("Info (%d): %v", key, i)
			}
		},
	}
)

func repoRequest(queryString string, client *graphqlclient.Client) (map[string]RepositoriesNodeList, error) {
	authStr := fmt.Sprintf("bearer %s", config.Token)

	req := graphqlclient.NewRequest(queryString)
	req.Var("org", config.Org)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)

	ctx := context.Background()

	var respData map[string]RepositoriesNodeList

	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, fmt.Errorf("graphql call: %w", err)
	}

	return respData, nil
}

func applySigning(repoSearchResult map[string]RepositoriesNodeList, client *graphqlclient.Client) (
	modified,
	info,
	errors []string,
) {
OUTER:
	for _, v := range repoSearchResult { // nolint

		// set for repositoryID
		modifyID := v.ID
		modifyFunc := createBranchProtection
		defaultBranch := v.DefaultBranchRef.Name

		if v.DefaultBranchRef.Name == "" {
			info = append(info, fmt.Sprintf("No default branch for %v", v.NameWithOwner))

			continue OUTER
		}

		// Check all nodes for default branch protection rule
		for _, node := range v.BranchProtectionRules.Nodes {
			if v.DefaultBranchRef.Name == node.Pattern {
				// set for branchProtectionRuleID
				modifyID = node.ID
				modifyFunc = updateBranchProtection

				// If default branch has already got signing turned on, no need to update
				if node.RequiresCommitSignatures {
					info = append(info, fmt.Sprintf("Signing already turned on for %v", v.NameWithOwner))

					continue OUTER
				}
			}
		}

		if err := modifyFunc(modifyID, defaultBranch, client); err != nil {
			errors = append(errors, err.Error())

			continue OUTER
		}

		modified = append(modified, v.NameWithOwner)
	}

	return modified, info, errors
}

func updateBranchProtection(branchProtectionID, branchName string, client *graphqlclient.Client) error {
	req := graphqlclient.NewRequest(`
		mutation UpdateBranchProtectionRule($branchProtectionId: String! $clientMutationId: String! $pattern: String!) {
			updateBranchProtectionRule(
				input:{
					clientMutationId: $clientMutationId,
					branchProtectionRuleId: $branchProtectionId,
					requiresCommitSignatures: true,
					pattern: $pattern,
				}
			) {
				clientMutationId
			}
		}
	`)
	req.Var("clientMutationId", fmt.Sprintf("github-tool-%v", branchProtectionID))
	req.Var("branchProtectionId", branchProtectionID)
	req.Var("pattern", branchName)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	ctx := context.Background()

	var respData interface{}

	if err := client.Run(ctx, req, &respData); err != nil {
		return fmt.Errorf("From API call: %w", err)
	}

	return nil
}

func createBranchProtection(repositoryID, branchName string, client *graphqlclient.Client) error {
	log.Printf("creating for branchName %+v", branchName)

	req := graphqlclient.NewRequest(`
		mutation CreateBranchProtectionRule($repositoryId: String! $clientMutationId: String! $pattern: String!) {
			createBranchProtectionRule(
				input:{
					clientMutationId: $clientMutationId,
					repositoryId: $repositoryId,
					requiresCommitSignatures: true,
					pattern: $pattern,
				}
			) {
				clientMutationId
			}
		}
	`)
	req.Var("clientMutationId", fmt.Sprintf("github-tool-%v", repositoryID))
	req.Var("repositoryId", repositoryID)
	req.Var("pattern", branchName)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	ctx := context.Background()

	var respData interface{}

	if err := client.Run(ctx, req, &respData); err != nil {
		return fmt.Errorf("From API call: %w", err)
	}

	return nil
}

func generateQuery(repos []string) string {
	preQueryStr := `
		fragment repoProperties on Repository {
			id
			nameWithOwner
			description
			defaultBranchRef {
				name
			}
			branchProtectionRules(first: 100) {
				nodes {
					id
					requiresCommitSignatures
					pattern
				}
			}
		}
		query ($org: String!) {
	`

	var signingQueryStr strings.Builder

	signingQueryStr.WriteString(preQueryStr)

	for i := 0; i < len(repos); i++ {
		signingQueryStr.WriteString(fmt.Sprintf(`
			repo%d: repository(owner: $org, name: "%s") {
				...repoProperties
		  	}
		`, i, repos[i]))
	}

	signingQueryStr.WriteString("}")

	return signingQueryStr.String()
}

// nolint // needed for cobra
func init() {
	signingCmd.Flags().StringVarP(&reposFile, "repos", "r", "", "file containing repositories on new line without org/ prefix. Max 100 repos")
	signingCmd.MarkFlagRequired("repos")
	rootCmd.AddCommand(signingCmd)
}

func readList(reposFile string) ([]string, error) {
	var repos []string

	file, err := os.Open(reposFile)
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
