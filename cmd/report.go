package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{ // nolint // needed for cobra
	Use:   "report",
	Short: "Run a report to generate a csv containing information on all organisation repos",
	RunE:  reportRun,
}

type report struct {
	reportGetter reportGetter
	reportCSV    reportCSV
	reportJSON   reportJSON
	reportAccess reportAccess
}

type reportGetter interface {
	getReport() ([]ReportResponse, error)
}

type reportGetterService struct{}

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

	filePath, err = cmd.Flags().GetString("file-path")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fileType, err = cmd.Flags().GetString("file-type")
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return reportCreate(
		&report{
			reportGetter: &reportGetterService{},
			reportCSV:    &reportCSVService{},
			reportJSON:   &reportJSONService{},
			reportAccess: &reportAccessService{},
		},
		dryRun,
		ignoreArchived,
		filePath,
		fileType,
	)
}

func reportCreate(r *report, dryRun, ignoreArchived bool, filePath, fileType string) error {
	allResults, err := r.reportGetter.getReport()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if dryRun {
		return nil
	}

	teamAccess, err := r.reportAccess.getReport()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if fileType == "json" {
		jsonReport, err := r.reportJSON.generate(ignoreArchived, allResults, teamAccess)
		if err != nil {
			return fmt.Errorf("generate json failed: %w", err)
		}

		if err := r.reportJSON.uploader(filePath, jsonReport); err != nil {
			return fmt.Errorf("upload json failed: %w", err)
		}

		return nil
	}

	lines := reportCSVGenerate(ignoreArchived, allResults, teamAccess)
	if err := reportCSVUpload(r.reportCSV, filePath, lines); err != nil {
		return fmt.Errorf("upload CSV failed: %w", err)
	}

	return nil
}

// nolint // needed for cobra
func init() {
	reportCmd.Flags().BoolVarP(&ignoreArchived, "ignore-archived", "i", true, "Ignore archived repositories")
	reportCmd.Flags().StringVarP(&filePath, "file-path", "f", "report.csv", "file path for report to be created, must be .csv or .json")
	reportCmd.Flags().StringVarP(&fileType, "file-type", "t", "csv", "file type, must be csv or json")
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
	query.WriteString("					id")
	query.WriteString("					deleteBranchOnMerge")
	query.WriteString("					isArchived")
	query.WriteString("					isEmpty")
	query.WriteString("					isFork")
	query.WriteString("					isPrivate")
	query.WriteString("					hasWikiEnabled")
	query.WriteString("					mergeCommitAllowed")
	query.WriteString("					name")
	query.WriteString("					nameWithOwner")
	query.WriteString("					rebaseMergeAllowed")
	query.WriteString("					squashMergeAllowed")
	query.WriteString("					branchProtectionRules(first: 100) {")
	query.WriteString("						nodes {")
	query.WriteString("							id")
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

func (r *reportGetterService) getReport() ([]ReportResponse, error) {
	var (
		cursor           *string
		totalRecordCount int
		allResults       []ReportResponse
		iteration        int
		bar              progressbar.Bar
	)

	client := graphqlclient.NewClient()

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

			return allResults, nil
		}

		if len(respData.Organization.Repositories.Nodes) > 0 {
			allResults = append(allResults, respData)
		}

		if iteration == 0 {
			bar.NewOption(0, totalRecordCount)
		}

		bar.Play(iteration)

		iteration += IterationCount

		if !respData.Organization.Repositories.PageInfo.HasNextPage {
			bar.Play(totalRecordCount)

			break
		}
	}

	bar.Finish("Get repository data")

	return allResults, nil
}
