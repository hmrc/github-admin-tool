package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/csv"
	"log"
	"strconv"

	"github.com/machinebox/graphql"
	"github.com/spf13/cobra"
)

var (
	dryRun         bool
	ignoreArchived bool
	reportCmd      = &cobra.Command{
		Use:   "report",
		Short: "Run a report to generate a csv containing information on all organisation repos",
		Run: func(cmd *cobra.Command, args []string) {
			dryRun, _ = cmd.Flags().GetBool("dry-run")
			ignoreArchived, _ = cmd.Flags().GetBool("ignore-archived")
			all := reportRequest()
			parsed := reportParse(all)
			csv.Run(parsed)
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
		Url           string `json:"url"`
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

func reportRequest() []Response {
	var allResults []Response
	client := graphql.NewClient("https://api.github.com/graphql")
	reqStr := getReportQuery()
	authStr := fmt.Sprintf("bearer %s", config.Client.Token)

	var cursor *string = nil
	loopCount := 0
	totalRecordCount := 0

	for loopCount <= totalRecordCount {
		req := graphql.NewRequest(reqStr)
		req.Var("org", config.Client.Org)
		req.Var("after", cursor)
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Authorization", authStr)

		ctx := context.Background()

		var respData Response
		if err := client.Run(ctx, req, &respData); err != nil {
			log.Fatal(err)
		}

		if !respData.Organization.Repositories.PageInfo.HasNextPage {
			break
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
	}
	return allResults
}

func reportParse(allResults []Response) [][]string {
	var parsed [][]string
	for _, allData := range allResults {
		for _, repo := range allData.Organization.Repositories.Nodes {
			if ignoreArchived && repo.IsArchived {
				continue
			}
			repoSlice := []string{
				repo.NameWithOwner,
				repo.DefaultBranchRef.Name,
				strconv.FormatBool(repo.IsArchived),
				strconv.FormatBool(repo.IsPrivate),
				strconv.FormatBool(repo.IsEmpty),
				strconv.FormatBool(repo.IsFork),
				repo.Parent.NameWithOwner,
				strconv.FormatBool(repo.MergeCommitAllowed),
				strconv.FormatBool(repo.SquashMergeAllowed),
				strconv.FormatBool(repo.RebaseMergeAllowed),
			}
			for _, protection := range repo.BranchProtectionRules.Nodes {
				repoSlice = append(repoSlice,
					strconv.FormatBool(protection.IsAdminEnforced),
					strconv.FormatBool(protection.RequiresCommitSignatures),
					strconv.FormatBool(protection.RestrictsPushes),
					strconv.FormatBool(protection.RequiresApprovingReviews),
					strconv.FormatBool(protection.RequiresStatusChecks),
					strconv.FormatBool(protection.RequiresCodeOwnerReviews),
					strconv.FormatBool(protection.DismissesStaleReviews),
					strconv.FormatBool(protection.RequiresStrictStatusChecks),
					strconv.Itoa(protection.RequiredApprovingReviewCount),
					strconv.FormatBool(protection.AllowsForcePushes),
					strconv.FormatBool(protection.AllowsDeletions),
					protection.Pattern,
				)
			}

			parsed = append(parsed, repoSlice)
		}
	}
	return parsed
}
