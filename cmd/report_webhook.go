package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"github-admin-tool/restclient"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var (
	reportWebookCmd = &cobra.Command{ // nolint // needed for cobra
		Use:   "report-webhook",
		Short: "Run a report to generate a csv containing webhooks for organisation repos",
		RunE:  reportWebookRun,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			fmt.Printf("{\"lastCursor\": \"%+v\", \"completedAllCalls\": %t}", lastCursorUsed, completedAllApiCalls)
		},
	}
	totalRestApiCalls = 0
	totalGraphqlCalls = 0
)

// nolint // needed for cobra
func init() {
	reportWebookCmd.Flags().BoolVarP(&ignoreArchived, "ignore-archived", "i", true, "Ignore archived repositores")
	reportWebookCmd.Flags().StringVarP(&filePath, "file-path", "f", "report.csv", "File path for report to be created, must be .csv")
	reportWebookCmd.Flags().StringVarP(&startCursor, "start-cursor", "s", "", "The starting cursor for webhook search to start from")
	rootCmd.AddCommand(reportWebookCmd)
}

type reportWebook struct {
	reportWebookGetter reportWebookGetter
	reportCSV          reportCSV
}

type reportWebookGetter interface {
	getRepositoryList(bool) ([]repositoryCursorList, error)
	getWebhooks([]repositoryCursorList) (map[string][]WebhookResponse, error)
}

type repositoryCursorList struct {
	cursor       string
	repositories []string
}

type reportWebookGetterService struct{}

func reportWebookRun(cmd *cobra.Command, args []string) error {
	var err error

	dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	ignoreArchived, err = cmd.Flags().GetBool("ignore-archived")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	filePath, err = cmd.Flags().GetString("file-path")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return reportWebookCreate(
		&reportWebook{
			reportWebookGetter: &reportWebookGetterService{},
			reportCSV:          &reportCSVService{},
		},
		dryRun,
		ignoreArchived,
		filePath,
	)
}

func reportWebookCreate(r *reportWebook, dryRun, ignoreArchived bool, filePath string) error {
	allRepositories, err := r.reportWebookGetter.getRepositoryList(ignoreArchived)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if !dryRun {
		allWebhooks, err := r.reportWebookGetter.getWebhooks(allRepositories)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		lines := reportCSVWebhookGenerate(allWebhooks)
		if err := r.reportCSV.uploader(filePath, lines); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
	}

	log.Printf("total rest api calls %d", totalRestApiCalls)
	log.Printf("total graphql calls %d", totalGraphqlCalls)

	return nil
}

func reportWebookQuery() string {
	var query strings.Builder

	query.WriteString("query ($org: String! $after: String) {")
	query.WriteString("		organization(login:$org) {")
	query.WriteString("			repositories(first: 100, after: $after, orderBy: {field: NAME, direction: ASC}) {")
	query.WriteString("				totalCount")
	query.WriteString("				pageInfo {")
	query.WriteString("					endCursor")
	query.WriteString("					hasNextPage")
	query.WriteString("				}")
	query.WriteString("				nodes {")
	query.WriteString("					isArchived")
	query.WriteString("					name")
	query.WriteString("				}")
	query.WriteString("			}")
	query.WriteString("		}")
	query.WriteString("}")

	return query.String()
}

func reportWebookRequest(queryString string) *graphqlclient.Request {
	authStr := fmt.Sprintf("bearer %s", config.Token)

	req := graphqlclient.NewRequest(queryString)
	req.Var("org", config.Org)

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)

	return req
}

func (r *reportWebookGetterService) getRepositoryList(ignoreArchived bool) ([]repositoryCursorList, error) {
	var (
		cursor     *string
		totalCount int
		result     []repositoryCursorList
		iteration  int
		bar        progressbar.Bar
	)

	client := graphqlclient.NewClient("https://api.github.com/graphql")
	query := reportWebookQuery()
	req := reportWebookRequest(query)
	ctx := context.Background()
	iteration = 0

	if startCursor != "" {
		cursor = &startCursor
	}

	for {
		// Set new cursor on every loop to paginate through 100 at a time
		req.Var("after", cursor)

		var response WebhookRepositoryResponse
		if err := client.Run(ctx, req, &response); err != nil {
			return result, fmt.Errorf("graphql call: %w", err)
		}

		cursor = &response.Organization.Repositories.PageInfo.EndCursor
		totalCount = response.Organization.Repositories.TotalCount

		if dryRun {
			log.Printf("This is a dry run, the report would process %d records\n", totalCount)

			return result, nil
		}

		repositoryList := []string{}
		for _, node := range response.Organization.Repositories.Nodes {
			if ignoreArchived && node.IsArchived {
				continue
			}

			repositoryList = append(repositoryList, node.Name)
		}

		result = append(result, repositoryCursorList{cursor: *cursor, repositories: repositoryList})

		if iteration == 0 {
			bar.NewOption(0, totalCount)
		}

		bar.Play(iteration)

		iteration += IterationCount
		totalGraphqlCalls++

		if !response.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
	}

	bar.Play(totalCount)
	bar.Finish("Get repository data")

	return result, nil
}

func (r *reportWebookGetterService) getWebhooks(repositories []repositoryCursorList) (map[string][]WebhookResponse, error) {
	allResults := make(map[string][]WebhookResponse, len(repositories))

	ctx := context.Background()
	totalCount := len(repositories)
	var bar progressbar.Bar

	iteration := 0

	// This loops through every set of 100 repos and set last cursor for each group
	for _, repositoryCursorList := range repositories {
		lastCursorUsed = repositoryCursorList.cursor

		if iteration == 0 {
			bar.NewOption(0, totalCount)
		}
		bar.Play(iteration)
		iteration++

		for _, repositoryName := range repositoryCursorList.repositories {

			client := restclient.NewClient(
				fmt.Sprintf("https://api.github.com/repos/%s/%s/hooks", config.Org, repositoryName),
				config.Token,
			)

			response := []WebhookResponse{}
			if err, statusCode := client.Run(ctx, &response); err != nil {
				//return allResults, fmt.Errorf("rest call: %w", err)

				// Assuming forbidden code is when rate limit is hit
				if statusCode == http.StatusForbidden {
					return allResults, nil
				}

				// Ignore any other errors and continue to top of loop
				log.Printf("Error: %s", err.Error())

				continue
			}

			allResults[repositoryName] = response
			totalRestApiCalls++
		}
	}

	bar.Play(totalCount)
	bar.Finish("Get webhook data")

	completedAllApiCalls = true

	return allResults, nil
}
