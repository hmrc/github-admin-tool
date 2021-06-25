package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"os"
	"regexp"
	"strings"
)

func repoList(reposFile string) ([]string, error) {
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

func repoQuery(repos []string) string {
	var signingQueryStr strings.Builder

	signingQueryStr.WriteString("fragment repoProperties on Repository {")
	signingQueryStr.WriteString("	id")
	signingQueryStr.WriteString("	nameWithOwner")
	signingQueryStr.WriteString("	description")
	signingQueryStr.WriteString("	defaultBranchRef {")
	signingQueryStr.WriteString("		name")
	signingQueryStr.WriteString("	}")
	signingQueryStr.WriteString("	branchProtectionRules(first: 100) {")
	signingQueryStr.WriteString("		nodes {")
	signingQueryStr.WriteString("			id")
	signingQueryStr.WriteString("			requiresCommitSignatures")
	signingQueryStr.WriteString("			pattern")
	signingQueryStr.WriteString("			requiresApprovingReviews")
	signingQueryStr.WriteString("			requiresCodeOwnerReviews")
	signingQueryStr.WriteString("			requiredApprovingReviewCount")
	signingQueryStr.WriteString("			dismissesStaleReviews")
	signingQueryStr.WriteString("		}")
	signingQueryStr.WriteString("	}")
	signingQueryStr.WriteString("}")
	signingQueryStr.WriteString("query ($org: String!) {")

	for i := 0; i < len(repos); i++ {
		signingQueryStr.WriteString(fmt.Sprintf("repo%d: repository(owner: $org, name: \"%s\") {", i, repos[i]))
		signingQueryStr.WriteString("	...repoProperties")
		signingQueryStr.WriteString("}")
	}

	signingQueryStr.WriteString("}")

	return signingQueryStr.String()
}

func repoRequest(queryString string, client *graphqlclient.Client) (map[string]*RepositoriesNode, error) {
	authStr := fmt.Sprintf("bearer %s", config.Token)

	req := graphqlclient.NewRequest(queryString)
	req.Var("org", config.Org)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)

	ctx := context.Background()

	var respData map[string]*RepositoriesNode

	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, fmt.Errorf("graphql call: %w", err)
	}

	return respData, nil
}
