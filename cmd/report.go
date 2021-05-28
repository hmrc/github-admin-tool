package cmd

import (
	"context"
	"fmt"
	"time"

	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"log"

	"github.com/spf13/cobra"
)

var (
	ignoreArchived bool
	reportCmd      = &cobra.Command{
		Use:   "report",
		Short: "Run a report to generate a csv containing information on all organisation repos",
		Run: func(cmd *cobra.Command, args []string) {
			client := graphqlclient.NewClient("https://api.github.com/graphql")

			var err error = nil
			dryRun, err = cmd.Flags().GetBool("dry-run")
			if err != nil {
				log.Fatal(err)
			}
			ignoreArchived, err = cmd.Flags().GetBool("ignore-archived")
			if err != nil {
				log.Fatal(err)
			}

			allResults := reportRequest(client)
			if !dryRun {
				GenerateCsv(ignoreArchived, allResults)
			}
		},
	}
)

func init() {
	reportCmd.Flags().BoolVarP(&ignoreArchived, "ignore-archived", "i", false, "Ignore archived repositores")
	rootCmd.AddCommand(reportCmd)
}

func reportRequest(client *graphqlclient.Client) []ReportResponse {
	reqStr := ReportQueryStr
	authStr := fmt.Sprintf("bearer %s", config.Client.Token)

	var (
		cursor           *string
		totalRecordCount int
		allResults       []ReportResponse
		iteration        int
		bar              progressbar.Bar
	)

	req := graphqlclient.NewRequest(reqStr)
	req.Var("org", config.Client.Org)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)
	ctx := context.Background()

	iteration = 0

	for {

		// Set new cursor on every loop to paginate through 100 at a time
		req.Var("after", cursor)

		var respData ReportResponse
		if err := client.Run(ctx, req, &respData); err != nil {
			log.Fatal(err)
		}

		cursor = &respData.Organization.Repositories.PageInfo.EndCursor
		totalRecordCount = respData.Organization.Repositories.TotalCount

		if dryRun {
			fmt.Printf("This is a dry run, the report would process %d records\n", totalRecordCount)

			break
		}

		// Set up progress bar
		if iteration == 0 {
			bar.NewOption(0, int64(totalRecordCount))
			if totalRecordCount <= 100 {
				iteration = totalRecordCount
			}
		} else if !respData.Organization.Repositories.PageInfo.HasNextPage {
			iteration = totalRecordCount
		}
		time.Sleep(100 * time.Millisecond)
		bar.Play(int64(iteration))

		if len(respData.Organization.Repositories.Nodes) > 0 {
			allResults = append(allResults, respData)
		}

		if !respData.Organization.Repositories.PageInfo.HasNextPage {
			break
		}

		iteration += 100
	}

	bar.Finish()

	return allResults
}
