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
	TeamPermissions       string                `json:"teamPermissions"`
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

type OrganizationTeams struct {
	Teams Teams `json:"teams"`
}

type OrganizationTeamsReponse struct {
	OrganizationTeams OrganizationTeams `json:"organization"`
}

type Teams struct {
	TeamNodes []TeamNodes `json:"nodes"`
}

type TeamNodes struct {
	TeamRepositories TeamRepositories `json:"repositories"`
}

type TeamRepositories struct {
	TotalCount int                `json:"totalCount"`
	PageInfo   PageInfo           `json:"pageInfo"`
	Edges      []RepositoriesEdge `json:"edges"`
}

type RepositoriesEdge struct {
	Node       RepositoriesEdgeNode `json:"node"`
	Permission string               `json:"permission"`
}

type RepositoriesEdgeNode struct {
	Name string `json:"name"`
}

type WebhookRepositoryResponse struct {
	Organization struct {
		Repositories struct {
			PageInfo struct {
				EndCursor   string `json:"endCursor"`
				HasNextPage bool   `json:"hasNextPage"`
			} `json:"pageInfo"`
			TotalCount int `json:"totalCount"`
			Nodes      []struct {
				Name       string `json:"name"`
				IsArchived bool   `json:"isArchived"`
			} `json:"nodes"`
		} `json:"repositories"`
	} `json:"organization"`
}

type WebhookResponse struct {
	Config struct {
		URL         string `json:"url"`
		InsecureURL int    `json:"insecure_url"` // nolint // this is from github
	} `json:"config"`
	Active bool     `json:"active"`
	ID     int      `json:"id"`
	Events []string `json:"events"`
}
