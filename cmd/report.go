package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"log"

	"github.com/spf13/cobra"
)

var (
	dryRun         bool
	ignoreArchived bool
	reportCmd      = &cobra.Command{
		Use:   "report",
		Short: "Run a report to generate a csv containing information on all organisation repos",
		Run: func(cmd *cobra.Command, args []string) {
			client := graphqlclient.NewClient("https://api.github.com/graphql")
			dryRun, _ = cmd.Flags().GetBool("dry-run")
			ignoreArchived, _ = cmd.Flags().GetBool("ignore-archived")
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

func reportRequest(client *graphqlclient.Client) []Response {
	reqStr := ReportQueryStr
	authStr := fmt.Sprintf("bearer %s", config.Client.Token)

	var (
		cursor           *string
		loopCount        int
		totalRecordCount int
		allResults       []Response
	)

	for loopCount <= totalRecordCount {
		req := graphqlclient.NewRequest(reqStr)
		req.Var("org", config.Client.Org)
		req.Var("after", cursor)
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Authorization", authStr)

		ctx := context.Background()

		var respData Response
		if err := client.Run(ctx, req, &respData); err != nil {
			log.Fatal(err)
		}

		cursor = &respData.Organization.Repositories.PageInfo.EndCursor
		totalRecordCount = respData.Organization.Repositories.TotalCount

		if dryRun {
			fmt.Printf("This is a dry run, the report would process %d records\n", totalRecordCount)
			break
		}

		loopCount += 100

		if len(respData.Organization.Repositories.Nodes) > 0 {
			fmt.Printf("Processing %d of %d total repos\n", loopCount, totalRecordCount)
			allResults = append(allResults, respData)
		}

		if !respData.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
	}

	return allResults
}
