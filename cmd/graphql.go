package cmd

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

type ReportResponse struct {
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

var ReportQueryStr string = `
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
