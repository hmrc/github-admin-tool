package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

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

			if len(repoMap) > 100 {
				log.Fatal(fmt.Errorf("Number of repos passed in (%d) must be less than 100", len(repoMap)))
			}

			queryString, err := generateQueryStr(repoMap)
			if err != nil {
				log.Fatal(err)
			}

			//fmt.Printf("queryString is %v", queryString) // nolint // TODO - REMOVE WHEN FINISHED building

			client := graphqlclient.NewClient("https://api.github.com/graphql")
			repoSearchResult, err := repoRequest(client, queryString)
			if err != nil {
				log.Fatal(err)
			}

			// fmt.Printf("repo to sign %v", repoSearchResult)

			signedRepos, err := applySigning(repoSearchResult)

			fmt.Printf("repo to sign %v", signedRepos)

			//fmt.Printf("repo list is %v", repos) // nolint // TODO - REMOVE WHEN FINISHED building
			// UpdateBranchProtectionRuleInput - requiresCommitSignatures
		},
	}
)

func repoRequest(client *graphqlclient.Client, queryString string) (map[string]RepositoriesNodeList, error) {
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

func applySigning(repoSearchResult map[string]RepositoriesNodeList) ([]string, error) {
	var (
		signedRepos []string
		err         error
	)

	// fmt.Printf("SigningResponse %+v", repoSearchResult)
	// var unmarshalled SigningResponse
	// json.Unmarshal(repoSearchResult.([]byte), &unmarshalled)
	// fmt.Printf("unmarshalled is %v", unmarshalled)

	// testrepo2-github-admin-tool - no bpr on default branch

	for _, v := range repoSearchResult {
		if v.DefaultBranchRef.Name == "" {
			// the_seven_good_dwarfs - no default branch, no bpr
			log.Printf("No default branch for %v", v.NameWithOwner)
			continue
		}

		if len(v.BranchProtectionRules.Nodes) == 0 {
			// testrepo1-github-admin-tool - no bpr
			log.Printf("No branch protection rules for %v", v.NameWithOwner)

			// Create branch protection
			// continue

		} else {
			// Check all nodes for default branch protection rule
			defaultBranchProtectionExists := false
			for _, node := range v.BranchProtectionRules.Nodes {
				if v.DefaultBranchRef.Name == node.Pattern {
					defaultBranchProtectionExists = true
				}
			}

			if !defaultBranchProtectionExists {
				// Create branch protection
				// continue
				log.Printf("No branch protection rules for default branch in repo %v with default branch %v", v.NameWithOwner, v.DefaultBranchRef.Name)

			}

			// Update branch protection
		}

		signedRepos = append(signedRepos, v.NameWithOwner)

	}

	return signedRepos, err
}

func updateBranchProtection(branchProtectionId string) (bool, error) {
	var (
		updated bool
		err     error
	)

	// mutation UpdateBranchProtectionRule {
	// 	updateBranchProtectionRule(input:{clientMutationId:"github_tool_id",branchProtectionRuleId:HOORAY}) {
	// 	  clientMutationId
	// 	}

	//   }

	return updated, err
}

// mutation CreateBranchProtectionRule {
// 	createBranchProtectionRule(input:{clientMutationId:"github_tool_id",branchProtectionRuleId:HOORAY}) {
// 	  clientMutationId
// 	}

//   }

func generateQueryStr(repos []string) (string, error) {
	preQueryStr := `
		fragment repoProperties on Repository {
			nameWithOwner
			description
			defaultBranchRef {
			name
			}
			branchProtectionRules(first: 100) {
				nodes {
					isAdminEnforced
					requiresCommitSignatures
					restrictsPushes
					requiresApprovingReviews
					requiresStatusChecks
					requiresCodeOwnerReviews
					dismissesStaleReviews
					requiresStrictStatusChecks
					requiredApprovingReviewCount
					allowsForcePushes
					allowsDeletions
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

	return signingQueryStr.String(), nil
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
