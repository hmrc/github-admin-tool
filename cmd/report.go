package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	dryRun         bool
	ignoreArchived bool
	allResults     []Response
	reportCmd      = &cobra.Command{
		Use:   "report",
		Short: "Run a report to generate a csv containing information on all organisation repos",
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ = cmd.Flags().GetBool("dry-run")
			ignoreArchived, _ = cmd.Flags().GetBool("ignore-archived")
			reportRequest()
			if !dryRun {
				GenerateCsv(ignoreArchived, allResults)
			}
		},
	}
)

type BranchProtectionRulesNodesList struct {
	IsAdminEnforced              bool   `json:"isAdminEnforced"`
	RequiresCommitSignatures     bool   `json:"requiresCommitSignatures"`
	RestrictsPushes              bool   `json:"restrictsPushes"`
	RequiresApprovingReviews     bool   `json:"requiresApprovingReviews"`
	RequiresStatusChecks         bool   `json:"requiresStatusChecks"`
	RequiresCodeOwnerReviews     bool   `json:"requiresCodeOwnerReviews"`
	DismissesStaleReviews        bool   `json:"dismissesStaleReviews"`
	RequiresStrictStatusChecks   bool   `json:"requiresStrictStatusChecks"`
	RequiredApprovingReviewCount int    `json:"requiredApprovingReviewCount"`
	AllowsForcePushes            bool   `json:"allowsForcePushes"`
	AllowsDeletions              bool   `json:"allowsDeletions"`
	Pattern                      string `json:"pattern"`
}

type RepositoriesNodeList struct {
	DeleteBranchOnMerge   bool   `json:"deleteBranchOnMerge"`
	IsArchived            bool   `json:"isArchived"`
	IsEmpty               bool   `json:"isEmpty"`
	IsFork                bool   `json:"isFork"`
	IsPrivate             bool   `json:"isPrivate"`
	MergeCommitAllowed    bool   `json:"mergeCommitAllowed"`
	Name                  string `json:"name"`
	NameWithOwner         string `json:"nameWithOwner"`
	RebaseMergeAllowed    bool   `json:"rebaseMergeAllowed"`
	SquashMergeAllowed    bool   `json:"squashMergeAllowed"`
	BranchProtectionRules struct {
		Nodes []BranchProtectionRulesNodesList `json:"nodes"`
	} `json:"branchProtectionRules"`
	Parent struct {
		Name          string `json:"name"`
		NameWithOwner string `json:"nameWithOwner"`
		URL           string `json:"url"`
	}
	DefaultBranchRef struct {
		Name string `json:"name"`
	}
}

type Response struct {
	Organization struct {
		Repositories struct {
			TotalCount int `json:"totalCount"`
			PageInfo   struct {
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
			} `json:"pageInfo"`
			Nodes []RepositoriesNodeList `json:"nodes"`
		} `json:"repositories"`
	} `json:"organization"`
}

func init() {
	reportCmd.Flags().BoolVarP(&ignoreArchived, "ignore-archived", "i", false, "Ignore archived repositores")
	rootCmd.AddCommand(reportCmd)
}

func getReportQuery() string {
	return `
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
    }`
}

func reportRequest() {
	client := NewClient("https://api.githubddd.com/graphql")
	reqStr := getReportQuery()
	authStr := fmt.Sprintf("bearer %s", config.Client.Token)

	var (
		cursor           *string
		loopCount        int
		totalRecordCount int
	)

	for loopCount <= totalRecordCount {
		req := NewRequest(reqStr)
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

		fmt.Printf("Processing %d of %d total repos\n", loopCount, totalRecordCount)
		allResults = append(allResults, respData)

		if !respData.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
	}
}
