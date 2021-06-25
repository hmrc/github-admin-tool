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

var (
	doReportRequest     = reportRequest     // nolint // Like this for testing mock
	doReportCSVGenerate = reportCSVGenerate // nolint // Like this for testing mock
	doRepositorySend    = repositorySend    // nolint // Like this for testing mock
	doRepositoryGet     = repositoryGet     // nolint // Like this for testing mock
)

func repositoryList(reposFile string) ([]string, error) {
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

func repositoryQuery(repos []string) string {
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

func repositoryRequest(queryString string) *graphqlclient.Request {
	authStr := fmt.Sprintf("bearer %s", config.Token)

	req := graphqlclient.NewRequest(queryString)
	req.Var("org", config.Org)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)

	return req
}

func repositorySend(req *graphqlclient.Request) (map[string]*RepositoriesNode, error) {
	ctx := context.Background()

	var respData map[string]*RepositoriesNode

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, fmt.Errorf("graphql call: %w", err)
	}

	return respData, nil
}

func repositoryGet(repositoryList []string) (repositories map[string]*RepositoriesNode, err error) {
	query := repositoryQuery(repositoryList)
	request := repositoryRequest(query)

	repositories, err = doRepositorySend(request)
	if err != nil {
		return repositories, fmt.Errorf("failure in repository get : %w", err)
	}

	return repositories, nil
}
