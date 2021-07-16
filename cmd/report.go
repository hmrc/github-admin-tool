package cmd

import (
	"context"
	"errors"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"log"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const (
	// IterationCount the number of repos per result set.
	IterationCount int = 100
)

var (
	ignoreArchived bool              // nolint // modifying within this package
	filePath       string            // nolint // modifying within this package
	reportCmd      = &cobra.Command{ // nolint // needed for cobra
		Use:   "report",
		Short: "Run a report to generate a csv containing information on all organisation repos",
		RunE:  reportRun,
	}
	doReportGet        = reportGet // nolint // Like this for testing mock
	errInvalidFilePath = errors.New("filepath must end with .csv")
)

func reportRun(cmd *cobra.Command, args []string) error {
	var err error

	dryRun, err = cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	ignoreArchived, err = cmd.Flags().GetBool("ignore-archived")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	allResults, err := doReportGet()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if !dryRun {
		filePath, err := cmd.Flags().GetString("file-path")
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		matched, err := regexp.MatchString(`.*.csv`, filePath)
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		if !matched {
			return errInvalidFilePath
		}

		if err = doReportCSVGenerate(filePath, ignoreArchived, allResults); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

// nolint // needed for cobra
func init() {
	reportCmd.Flags().BoolVarP(&ignoreArchived, "ignore-archived", "i", false, "Ignore archived repositores")
	reportCmd.Flags().StringVarP(&filePath, "file-path", "f", "report.csv", "file path for report to be created, must be .csv")
	rootCmd.AddCommand(reportCmd)
}

func reportQuery() string {
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
	query.WriteString("					deleteBranchOnMerge")
	query.WriteString("					isArchived")
	query.WriteString("					isEmpty")
	query.WriteString("					isFork")
	query.WriteString("					isPrivate")
	query.WriteString("					mergeCommitAllowed")
	query.WriteString("					name")
	query.WriteString("					nameWithOwner")
	query.WriteString("					rebaseMergeAllowed")
	query.WriteString("					squashMergeAllowed")
	query.WriteString("					branchProtectionRules(first: 100) {")
	query.WriteString("						nodes {")
	query.WriteString("							isAdminEnforced")
	query.WriteString("							requiresCommitSignatures")
	query.WriteString("							restrictsPushes")
	query.WriteString("							requiresApprovingReviews")
	query.WriteString("							requiresStatusChecks")
	query.WriteString("							requiresCodeOwnerReviews")
	query.WriteString("							dismissesStaleReviews")
	query.WriteString("							requiresStrictStatusChecks")
	query.WriteString("							requiredApprovingReviewCount")
	query.WriteString("							allowsForcePushes")
	query.WriteString("							allowsDeletions")
	query.WriteString("							pattern")
	query.WriteString("						}")
	query.WriteString("					}")
	query.WriteString("					defaultBranchRef {")
	query.WriteString("						name")
	query.WriteString("					}")
	query.WriteString("					parent {")
	query.WriteString("						name")
	query.WriteString("						nameWithOwner")
	query.WriteString("						url")
	query.WriteString("					}")
	query.WriteString("				}")
	query.WriteString("			}")
	query.WriteString("		}")
	query.WriteString("	}")

	return query.String()
}

func reportRequest(queryString string) *graphqlclient.Request {
	authStr := fmt.Sprintf("bearer %s", config.Token)

	req := graphqlclient.NewRequest(queryString)
	req.Var("org", config.Org)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)

	return req
}

func reportGet() ([]ReportResponse, error) {
	var (
		cursor           *string
		totalRecordCount int
		allResults       []ReportResponse
		iteration        int
		bar              progressbar.Bar
	)

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	query := reportQuery()

	req := reportRequest(query)

	ctx := context.Background()
	iteration = 0

	for {
		// Set new cursor on every loop to paginate through 100 at a time
		req.Var("after", cursor)

		var respData ReportResponse
		if err := client.Run(ctx, req, &respData); err != nil {
			return allResults, fmt.Errorf("graphql call: %w", err)
		}

		cursor = &respData.Organization.Repositories.PageInfo.EndCursor
		totalRecordCount = respData.Organization.Repositories.TotalCount

		if dryRun {
			log.Printf("This is a dry run, the report would process %d records\n", totalRecordCount)

			break
		}

		if len(respData.Organization.Repositories.Nodes) > 0 {
			allResults = append(allResults, respData)
		}

		if iteration == 0 {
			bar.NewOption(0, totalRecordCount)
		}

		if !respData.Organization.Repositories.PageInfo.HasNextPage {
			iteration = totalRecordCount
			bar.Play(iteration)

			break
		}

		bar.Play(iteration)

		iteration += IterationCount
	}

	bar.Finish()

	return allResults, nil
}
