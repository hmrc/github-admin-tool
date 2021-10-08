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

type repository struct {
	reader repositoryReader
	getter repositoryGetter
}

type repositoryReader interface {
	read(reposFile string) ([]string, error)
}

type repositoryReaderService struct{}

func (r *repositoryReaderService) read(reposFile string) ([]string, error) {
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

type repositoryGetter interface {
	get(repositoryList []string, sender *githubRepositorySender) (repositories map[string]*RepositoriesNode, err error)
}

type repositoryGetterService struct{}

func (r *repositoryGetterService) get(
	repositoryList []string,
	sender *githubRepositorySender,
) (
	repositories map[string]*RepositoriesNode,
	err error,
) {
	query := repositoryQuery(repositoryList)
	request := repositoryRequest(query)

	repositories, err = sender.sender.send(request)
	if err != nil {
		return repositories, fmt.Errorf("failure in repository get : %w", err)
	}

	return repositories, nil
}

type githubRepositorySender struct {
	sender repositorySender
}

type repositorySender interface {
	send(req *graphqlclient.Request) (map[string]*RepositoriesNode, error)
}

type repositorySenderService struct{}

func (r *repositorySenderService) send(req *graphqlclient.Request) (map[string]*RepositoriesNode, error) {
	ctx := context.Background()

	var respData map[string]*RepositoriesNode

	client := graphqlclient.NewClient()

	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, fmt.Errorf("graphql call: %w", err)
	}

	return respData, nil
}

func repositoryQuery(repos []string) string {
	var query strings.Builder

	query.WriteString("fragment repoProperties on Repository {")
	query.WriteString("	id")
	query.WriteString("	nameWithOwner")
	query.WriteString("	description")
	query.WriteString("	defaultBranchRef {")
	query.WriteString("		name")
	query.WriteString("	}")
	query.WriteString("	branchProtectionRules(first: 100) {")
	query.WriteString("		nodes {")
	query.WriteString("			id")
	query.WriteString("			requiresCommitSignatures")
	query.WriteString("			pattern")
	query.WriteString("			requiresApprovingReviews")
	query.WriteString("			requiresCodeOwnerReviews")
	query.WriteString("			requiredApprovingReviewCount")
	query.WriteString("			dismissesStaleReviews")
	query.WriteString("		}")
	query.WriteString("	}")
	query.WriteString("}")
	query.WriteString("query ($org: String!) {")

	for key, repositoryName := range repos {
		query.WriteString(fmt.Sprintf("repo%d: repository(owner: $org, name: \"%s\") {", key, repositoryName))
		query.WriteString("	...repoProperties")
		query.WriteString("}")
	}

	query.WriteString("}")

	return query.String()
}

func repositoryRequest(queryString string) *graphqlclient.Request {
	authStr := fmt.Sprintf("bearer %s", config.Token)

	req := graphqlclient.NewRequest(queryString)
	req.Var("org", config.Org)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)

	return req
}
