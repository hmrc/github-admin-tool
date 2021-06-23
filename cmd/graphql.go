package cmd

type BranchProtectionRulesNode struct {
	ID                           string `json:"id"`
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

type BranchProtectionRules struct {
	Nodes []BranchProtectionRulesNode `json:"nodes"`
}

type Parent struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
	URL           string `json:"url"`
}

type DefaultBranchRef struct {
	Name string `json:"name"`
}

type RepositoriesNode struct {
	ID                    string                `json:"id"`
	DeleteBranchOnMerge   bool                  `json:"deleteBranchOnMerge"`
	IsArchived            bool                  `json:"isArchived"`
	IsEmpty               bool                  `json:"isEmpty"`
	IsFork                bool                  `json:"isFork"`
	IsPrivate             bool                  `json:"isPrivate"`
	MergeCommitAllowed    bool                  `json:"mergeCommitAllowed"`
	Name                  string                `json:"name"`
	NameWithOwner         string                `json:"nameWithOwner"`
	RebaseMergeAllowed    bool                  `json:"rebaseMergeAllowed"`
	SquashMergeAllowed    bool                  `json:"squashMergeAllowed"`
	BranchProtectionRules BranchProtectionRules `json:"branchProtectionRules"`
	Parent                Parent
	DefaultBranchRef      DefaultBranchRef
}

type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type Repositories struct {
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
	Nodes      []RepositoriesNode `json:"nodes"`
}

type Organization struct {
	Repositories Repositories `json:"repositories"`
}

type ReportResponse struct {
	Organization Organization `json:"organization"`
}
