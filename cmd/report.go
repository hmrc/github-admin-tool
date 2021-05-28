package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"log"
	"time"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	// IterationCount the number of repos per result set.
	IterationCount        int   = 100
	MillisecondMultiplier int64 = 100
)

var (
	ignoreArchived bool
	reportCmd      = &cobra.Command{
		Use:   "report",
		Short: "Run a report to generate a csv containing information on all organisation repos",
		Run: func(cmd *cobra.Command, args []string) {
			client := graphqlclient.NewClient("https://api.github.com/graphql")

			var err error
			dryRun, err = cmd.Flags().GetBool("dry-run")
			if err != nil {
				log.Fatal(err)
			}
			ignoreArchived, err = cmd.Flags().GetBool("ignore-archived")
			if err != nil {
				log.Fatal(err)
			}

			allResults, err := reportRequest(client)
			if err != nil {
				log.Fatal(err)
			}
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

func reportRequest(client *graphqlclient.Client) ([]ReportResponse, error) {
	authStr := fmt.Sprintf("bearer %s", config.Client.Token)

	var (
		cursor           *string
		totalRecordCount int
		allResults       []ReportResponse
		iteration        int
		bar              progressbar.Bar
	)

	req := graphqlclient.NewRequest(ReportQueryStr)
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
			return allResults, errors.Wrap(err, "graphql call")
		}

		cursor = &respData.Organization.Repositories.PageInfo.EndCursor
		totalRecordCount = respData.Organization.Repositories.TotalCount

		if dryRun {
			log.Printf("This is a dry run, the report would process %d records\n", totalRecordCount)

			break
		}

		// Set up progress bar
		if iteration == 0 {
			bar.NewOption(0, int64(totalRecordCount))

			if totalRecordCount <= IterationCount {
				iteration = totalRecordCount
			}
		} else if !respData.Organization.Repositories.PageInfo.HasNextPage {
			iteration = totalRecordCount
		}

		time.Sleep(time.Millisecond * time.Duration(IterationCount))
		bar.Play(int64(iteration))

		if len(respData.Organization.Repositories.Nodes) > 0 {
			allResults = append(allResults, respData)
		}

		if !respData.Organization.Repositories.PageInfo.HasNextPage {
			break
		}

		iteration += IterationCount
	}

	bar.Finish()

	return allResults, nil
}
