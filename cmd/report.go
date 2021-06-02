package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"log"

	"github.com/spf13/cobra"
)

const (
	// IterationCount the number of repos per result set.
	IterationCount int = 100
)

var (
	ignoreArchived bool              // nolint // modifying within this package
	reportCmd      = &cobra.Command{ // nolint // needed for cobra
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
				if err = GenerateCSV(ignoreArchived, allResults); err != nil {
					log.Fatal(err)
				}
			}
		},
	}
)

// nolint // needed for cobra
func init() {
	reportCmd.Flags().BoolVarP(&ignoreArchived, "ignore-archived", "i", false, "Ignore archived repositores")
	rootCmd.AddCommand(reportCmd)
}

func reportRequest(client *graphqlclient.Client) ([]ReportResponse, error) {
	var (
		cursor           *string
		totalRecordCount int
		allResults       []ReportResponse
		iteration        int
		bar              progressbar.Bar
	)

	authStr := fmt.Sprintf("bearer %s", config.Client.Token)

	reportQueryStr := `
		query ($org: String! $after: String) {
			organization(login:$org) {
				repositories(first: 100, after: $after, orderBy: {field: NAME, direction: ASC}) {
					totalCount
					pageInfo {
						endCursor
						hasNextPage
					}
					nodes {
						deleteBranchOnMerge
						isArchived
						isEmpty
						isFork
						isPrivate
						mergeCommitAllowed
						name
						nameWithOwner
						rebaseMergeAllowed
						squashMergeAllowed
						branchProtectionRules(first: 100) {
							nodes {
								isAdminEnforced
								requiresCommitSignatures
								restrictsPushes
								requiresApprovingReviews
								requiresStatusChecks
								requiresCodeOwnerReviews
								dismissesStaleReviews
								requiresStrictStatusChecks
								requiredApprovingReviewCount
								allowsForcePushes
								allowsDeletions
								pattern
							}
						}
						defaultBranchRef {
							name
						}
						parent {
							name
							nameWithOwner
							url
						}
					}
				}
			}
		}
	`

	req := graphqlclient.NewRequest(reportQueryStr)
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
