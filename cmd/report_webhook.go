package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"github-admin-tool/ratelimit"
	"github-admin-tool/restclient"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	reportWebhookResponse WebhookCmdResponse // nolint // needed for cobra
	reportWebookCmd       = &cobra.Command{  // nolint // needed for cobra
		Use:   "report-webhook",
		Short: "Run a report to generate a csv containing webhooks for organisation repos",
		Long: `Webhook report can often run over 15 minutes depending on large number of repositories in your org.  
Use the timeout flag and resulting $file-path.status file to run again from cursor point if needed, 
this is useful when calling from a Lambda.`,
		RunE: reportWebookRun,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			response, err := json.Marshal(reportWebhookResponse)
			if err != nil {
				log.Print(fmt.Errorf("%w", err))
			}
			log.Printf("End time %v", time.Now().Format(time.RFC1123))

			jsonService := reportJSONService{}
			if err := jsonService.uploader(filePath+".status", response); err != nil {
				log.Print(fmt.Errorf("%w", err))
			}
		},
	}
)

func init() { // nolint // needed for cobra
	reportWebookCmd.Flags().BoolP("ignore-archived", "i", true, "Ignore archived repositores")
	reportWebookCmd.Flags().StringP(
		"file-path", "f", "report.csv", "File path for report to be created, must be .csv or .json",
	)
	reportWebookCmd.Flags().StringP("file-type", "t", "csv", "file type, must be csv or json")
	reportWebookCmd.Flags().StringP("start-cursor", "s", "", "The starting cursor for webhook search to start from")
	reportWebookCmd.Flags().IntP(
		"timeout", "o", 60, "Timeout for script (in minutes), useful when calling from Lambdas",
	)
	rootCmd.AddCommand(reportWebookCmd)
}

type WebhookCmdResponse struct {
	LastCursor         string
	CompletedAllCalls  bool
	RestCalls          int
	GraphqlCalls       int
	Errors             []string
	StartTimeSecs      int64
	EndTimeSecs        int64
	RateLimit          int
	RateLimitResetSecs int64
}

type reportWebook struct {
	reportWebookGetter reportWebookGetter
	reportCSV          reportCSV
	reportJSON         reportJSON
	dryRun             bool
	ignoreArchived     bool
	filePath           string
	fileType           string
	startCursor        string
	timeout            int
}

type reportWebookGetter interface {
	getRepositoryList(*reportWebook) ([]repositoryCursorList, error)
	getWebhooks(*reportWebook, []repositoryCursorList) (map[string][]WebhookResponse, error)
}

type repositoryCursorList struct {
	cursor       string
	repositories []string
}

type reportWebookGetterService struct{}

func reportWebookRun(cmd *cobra.Command, args []string) error {
	report := &reportWebook{
		reportWebookGetter: &reportWebookGetterService{},
		reportCSV:          &reportCSVService{},
		reportJSON:         &reportJSONService{},
	}

	if err := reportWebhookValidateFlags(report, cmd); err != nil {
		return err
	}

	setTimeout(report)

	if err := setRateLimit(); err != nil {
		return err
	}

	log.Printf("Rate limit remaining %d", reportWebhookResponse.RateLimit)
	log.Printf("Rate limit reset %v", time.Unix(reportWebhookResponse.RateLimitResetSecs, 0).Format(time.RFC1123))

	return reportWebookCreate(report)
}

func reportWebhookValidateFlags(r *reportWebook, cmd *cobra.Command) error {
	var err error

	r.dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	r.ignoreArchived, err = cmd.Flags().GetBool("ignore-archived")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	r.filePath, err = cmd.Flags().GetString("file-path")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	r.fileType, err = cmd.Flags().GetString("file-type")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	r.startCursor, err = cmd.Flags().GetString("start-cursor")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	r.timeout, err = cmd.Flags().GetInt("timeout")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if r.timeout > 60 || r.timeout < 1 {
		return errInvalidTimeout
	}

	return nil
}

func reportWebookCreate(r *reportWebook) error {
	allRepositories, err := r.reportWebookGetter.getRepositoryList(r)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if dryRun {
		return nil
	}

	allWebhooks, err := r.reportWebookGetter.getWebhooks(r, allRepositories)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if fileType == "json" {
		jsonReport, err := r.reportJSON.generateWebhook(allWebhooks)
		if err != nil {
			return fmt.Errorf("generate json failed: %w", err)
		}

		if err := r.reportJSON.uploader(filePath, jsonReport); err != nil {
			return fmt.Errorf("upload json failed: %w", err)
		}

		return nil
	}

	lines := reportCSVWebhookGenerate(allWebhooks)
	if err := reportCSVUpload(r.reportCSV, filePath, lines); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

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

func (r *reportWebookGetterService) getRepositoryList(report *reportWebook) ([]repositoryCursorList, error) {
	var (
		cursor     *string
		totalCount int
		result     []repositoryCursorList
		iteration  int
		bar        progressbar.Bar
	)

	client := graphqlclient.NewClient()
	query := reportWebookQuery()
	req := reportWebookRequest(query)
	ctx := context.Background()
	iteration = 0

	if report.startCursor != "" {
		cursor = &report.startCursor
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
		reportWebhookResponse.GraphqlCalls++

		if !response.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
	}

	bar.Play(totalCount)
	bar.Finish("Get repository data")

	return result, nil
}

func (r *reportWebookGetterService) getWebhooks(
	report *reportWebook,
	repositories []repositoryCursorList,
) (map[string][]WebhookResponse, error) {
	allResults := make(map[string][]WebhookResponse, len(repositories))

	ctx := context.Background()
	totalCount := len(repositories)
	iteration := 0

	var bar progressbar.Bar

	// This loops through every set of 100 repos and set last cursor for each group
	for _, repositoryCursorList := range repositories {
		if iteration == 0 {
			bar.NewOption(0, totalCount)
		}

		bar.Play(iteration)
		iteration++

		if hasReachedRateLimit() || hasTimeoutElapsed() {
			bar.Finish("Get webhook data timed out or rate limit hit")

			return allResults, nil
		}

		for _, repositoryName := range repositoryCursorList.repositories {
			client := restclient.NewClient(
				fmt.Sprintf("/repos/%s/%s/hooks", config.Org, repositoryName),
				config.Token,
			)

			response := []WebhookResponse{}
			if err := client.Run(ctx, &response); err != nil {
				// Ignore any other errors and continue to top of loop
				reportWebhookResponse.Errors = append(reportWebhookResponse.Errors, err.Error())

				continue
			}

			allResults[repositoryName] = response
			reportWebhookResponse.RestCalls++
		}

		reportWebhookResponse.LastCursor = repositoryCursorList.cursor
	}

	bar.Play(totalCount)
	bar.Finish("Get webhook data")

	reportWebhookResponse.CompletedAllCalls = true

	return allResults, nil
}

func setTimeout(r *reportWebook) {
	now := time.Now()
	reportWebhookResponse.StartTimeSecs = now.Unix()
	log.Printf("Start time %v", now.Format(time.RFC1123))

	reportWebhookResponse.EndTimeSecs = now.Add(time.Minute * time.Duration(r.timeout)).Unix()
}

func hasTimeoutElapsed() bool {
	currentSeconds := time.Now().Unix()

	return currentSeconds >= reportWebhookResponse.EndTimeSecs
}

func setRateLimit() error {
	rateResponse, err := ratelimit.GetRateLimit(config.Token)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	reportWebhookResponse.RateLimit = rateResponse.Resources.Rest.Remaining
	reportWebhookResponse.RateLimitResetSecs = rateResponse.Resources.Rest.Reset

	return nil
}

func hasReachedRateLimit() bool {
	// Reset rate limit
	if err := setRateLimit(); err != nil {
		reportWebhookResponse.Errors = append(reportWebhookResponse.Errors, err.Error())

		return true
	}

	// If limit cannot fulfill number in the next iteration then fail
	return reportWebhookResponse.RateLimit < IterationCount
}
