package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github-admin-tool/graphqlclient"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const maxRepositories = 100

var (
	reposFile  string            // nolint // global flag for cobra
	signingCmd = &cobra.Command{ // nolint // needed for cobra
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

	signingCreate  = createBranchProtection // nolint // Like this for testing mock
	signingUpdate  = updateBranchProtection // nolint // Like this for testing mock
	errInvalidRepo = errors.New("invalid repo name")
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

			if err = signingUpdate(branchProtection.ID, client); err != nil {
				problems = append(problems, err.Error())

				continue OUTER
			}
			modified = append(modified, repository.NameWithOwner)

			continue OUTER
		}

		if err = signingCreate(repository.ID, repository.DefaultBranchRef.Name, client); err != nil {
			problems = append(problems, err.Error())

			continue OUTER
		}

		created = append(created, repository.NameWithOwner)
	}

	return modified, created, info, problems
}

func updateBranchProtection(branchProtectionID string, client *graphqlclient.Client) error {
	req := graphqlclient.NewRequest(`
		mutation UpdateBranchProtectionRule($branchProtectionId: String! $clientMutationId: String!) {
			updateBranchProtectionRule(
				input:{
					clientMutationId: $clientMutationId,
					branchProtectionRuleId: $branchProtectionId,
					requiresCommitSignatures: true,
				}
			) {
				branchProtectionRule {
					id
				}
			}
		}
	`)
	req.Var("clientMutationId", fmt.Sprintf("github-tool-%v", branchProtectionID))
	req.Var("branchProtectionId", branchProtectionID)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	ctx := context.Background()

	if err := client.Run(ctx, req, nil); err != nil {
		return fmt.Errorf("from API call: %w", err)
	}

	return nil
}

func createBranchProtection(repositoryID, branchName string, client *graphqlclient.Client) error {
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
				branchProtectionRule {
					id
				}
			}
		}
	`)
	req.Var("clientMutationId", fmt.Sprintf("github-tool-%v", repositoryID))
	req.Var("repositoryId", repositoryID)
	req.Var("pattern", branchName)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	ctx := context.Background()

	if err := client.Run(ctx, req, nil); err != nil {
		return fmt.Errorf("from API call: %w", err)
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

	validRepoName := regexp.MustCompile("^[A-Za-z0-9_.-]+$")

	file, err := os.Open(reposFile)
	if err != nil {
		return repos, fmt.Errorf("could not open repo file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repoName := scanner.Text()
		if !validRepoName.MatchString(repoName) {
			return repos, fmt.Errorf("%w: %s", errInvalidRepo, repoName)
		}

		repos = append(repos, repoName)
	}

	return repos, nil
}
