package cmd

import (
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
)

func Test_branchProtectionQueryBlocks(t *testing.T) {
	type args struct {
		args []BranchProtectionArgs
	}

	tests := []struct {
		name            string
		args            args
		wantMutation    string
		wantInput       string
		wantRequestVars map[string]interface{}
	}{
		{
			name: "branchProtectionQueryBlocks",
			args: args{
				[]BranchProtectionArgs{{
					Name:     "requiresApprovingReviews",
					DataType: "Boolean",
					Value:    true,
				}},
			},
			wantMutation:    "$requiresApprovingReviews: Boolean!,",
			wantInput:       "requiresApprovingReviews: $requiresApprovingReviews,",
			wantRequestVars: map[string]interface{}{"requiresApprovingReviews": true},
		},
		{
			name: "branchProtectionQueryBlocks with more than one arg",
			args: args{
				[]BranchProtectionArgs{
					{
						Name:     "requiresApprovingReviews",
						DataType: "Boolean",
						Value:    true,
					},
					{
						Name:     "requiredApprovingReviewCount",
						DataType: "Int",
						Value:    5,
					},
				},
			},
			wantMutation: "$requiresApprovingReviews: Boolean!,$requiredApprovingReviewCount: Int!,",
			wantInput: "requiresApprovingReviews: $requiresApprovingReviews," +
				"requiredApprovingReviewCount: $requiredApprovingReviewCount,",
			wantRequestVars: map[string]interface{}{"requiresApprovingReviews": true, "requiredApprovingReviewCount": 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMutation, gotInput, gotRequestVars := branchProtectionQueryBlocks(tt.args.args)
			if !reflect.DeepEqual(gotMutation, tt.wantMutation) {
				t.Errorf("branchProtectionQueryBlocks() gotMutation = %v, want %v", gotMutation, tt.wantMutation)
			}
			if !reflect.DeepEqual(gotInput, tt.wantInput) {
				t.Errorf("branchProtectionQueryBlocks() gotInput = %v, want = %v", gotInput, tt.wantInput)
			}
			if !reflect.DeepEqual(gotRequestVars, tt.wantRequestVars) {
				t.Errorf("branchProtectionQueryBlocks() gotRequestVars = %v, want %v", gotRequestVars, tt.wantRequestVars)
			}
		})
	}
}

var (
	errMockSend   = errors.New("from API call: something went wrong")
	errMockCreate = errors.New("create: test")
	errMockUpdate = errors.New("update: test")
)

func mockDoBranchProtectionSend(req *graphqlclient.Request, client *graphqlclient.Client) error {
	return nil
}

func mockDoBranchProtectionSendError(req *graphqlclient.Request, client *graphqlclient.Client) error {
	return errMockSend
}

func mockDoBranchProtectionUpdate(branchProtectionArgs []BranchProtectionArgs, branchProtectionRuleID string) error {
	return nil
}

func mockDoBranchProtectionUpdateError(
	branchProtectionArgs []BranchProtectionArgs,
	branchProtectionRuleID string,
) error {
	return errMockUpdate
}

func mockDoBranchProtectionCreate(branchProtectionArgs []BranchProtectionArgs, repoID, pattern string) error {
	return nil
}

func mockDoBranchProtectionCreateError(branchProtectionArgs []BranchProtectionArgs, repoID, pattern string) error {
	return errMockCreate
}

func Test_branchProtectionApply(t *testing.T) {
	type args struct {
		repoSearchResult     map[string]*RepositoriesNode
		action               string
		branchProtectionArgs []BranchProtectionArgs
	}

	doBranchProtectionUpdate = mockDoBranchProtectionUpdate
	defer func() { doBranchProtectionUpdate = branchProtectionUpdate }()

	doBranchProtectionCreate = mockDoBranchProtectionCreate
	defer func() { doBranchProtectionCreate = branchProtectionCreate }()

	tests := []struct {
		name               string
		args               args
		wantModified       []string
		wantCreated        []string
		wantInfo           []string
		wantErrors         []string
		mockErrorFunctions bool
	}{
		{
			name: "branchProtectionApply with no default branch",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/some-repo-name",
				}},
			},
			wantModified: nil,
			wantCreated:  nil,
			wantInfo:     []string{"No default branch for org/some-repo-name"},
			wantErrors:   nil,
		},
		{
			name: "branchProtectionApply with no default branch protection rule",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/no-branch-protection",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNode{{
							RequiresCommitSignatures: true,
							Pattern:                  "another-branch-name",
						}},
					},
				}},
			},
			wantModified: nil,
			wantCreated:  []string{"org/no-branch-protection"},
			wantInfo:     nil,
			wantErrors:   nil,
		},
		{
			name: "branchProtectionApply with default branch protection rule signing on",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/signing-on",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNode{{
							RequiresCommitSignatures: true,
							Pattern:                  "default-branch-name",
						}},
					},
				}},
				action: "Signing",
			},
			wantModified: nil,
			wantCreated:  nil,
			wantInfo:     []string{"Signing already turned on for org/signing-on"},
			wantErrors:   nil,
		},
		{
			name: "branchProtectionApply with default branch protection rule signing off",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/signing-off",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNode{{
							RequiresCommitSignatures: false,
							Pattern:                  "default-branch-name",
						}},
					},
				}},
				action: "Signing",
			},
			wantModified: []string{"org/signing-off"},
			wantCreated:  nil,
			wantInfo:     nil,
			wantErrors:   nil,
		},
		{
			name: "branchProtectionApply creating failure",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/signing-off",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
				}},
			},
			wantModified:       nil,
			wantCreated:        nil,
			wantInfo:           nil,
			wantErrors:         []string{"create: test"},
			mockErrorFunctions: true,
		},
		{
			name: "branchProtectionApply with default branch protection rule pr approval settings the same",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/pr-approval-duplicate",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNode{{
							RequiresApprovingReviews:     true,
							RequiredApprovingReviewCount: 1,
							DismissesStaleReviews:        true,
							RequiresCodeOwnerReviews:     false,
							Pattern:                      "default-branch-name",
						}},
					},
				}},
				action: "Pr-approval",
			},
			wantModified: nil,
			wantCreated:  nil,
			wantInfo:     []string{"Pr-approval settings already set for org/pr-approval-duplicate"},
			wantErrors:   nil,
		},
		{
			name: "branchProtectionApply pr approval update failure",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/pr-approval-test",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNode{{
							RequiresCommitSignatures: true,
							Pattern:                  "default-branch-name",
						}},
					},
				}},
				action: "Pr-approval",
			},
			wantErrors:         []string{"update: test"},
			mockErrorFunctions: true,
		},
		{
			name: "branchProtectionApply pr approval create failure",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/pr-approval-test",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
				}},
				action: "Pr-approval",
			},
			wantErrors:         []string{"create: test"},
			mockErrorFunctions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErrorFunctions {
				doBranchProtectionUpdate = mockDoBranchProtectionUpdateError
				defer func() { doBranchProtectionUpdate = branchProtectionUpdate }()

				doBranchProtectionCreate = mockDoBranchProtectionCreateError
				defer func() { doBranchProtectionCreate = branchProtectionCreate }()
			}

			gotModified, gotCreated, gotInfo, gotErrors := branchProtectionApply(
				tt.args.repoSearchResult,
				tt.args.action,
				tt.args.branchProtectionArgs,
			)
			if !reflect.DeepEqual(gotModified, tt.wantModified) {
				t.Errorf("applySigning() gotModified = %v, want %v", gotModified, tt.wantModified)
			}
			if !reflect.DeepEqual(gotCreated, tt.wantCreated) {
				t.Errorf("applySigning() gotCreated = %v, want %v", gotCreated, tt.wantCreated)
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("applySigning() gotInfo = %v, want %v", gotInfo, tt.wantInfo)
			}
			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
				t.Errorf("applySigning() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
			}
		})
	}
}

func Test_branchProtectionQuery(t *testing.T) {
	type args struct {
		branchProtectionArgs []BranchProtectionArgs
		action               string
	}

	tests := []struct {
		name            string
		args            args
		filePath        string
		wantRequestVars map[string]interface{}
	}{
		{
			name: "branchProtectionQuery returns update query",
			args: args{
				branchProtectionArgs: []BranchProtectionArgs{
					{
						Name:     "requiresApprovingReviews",
						DataType: "Boolean",
						Value:    true,
					},
					{
						Name:     "branchProtectionRuleId",
						DataType: "String",
						Value:    "some-rule-id",
					},
				},
				action: "update",
			},
			filePath: "../testdata/mockUpdateBranchProtectionQuery.txt",
			wantRequestVars: map[string]interface{}{
				"branchProtectionRuleId":   "some-rule-id",
				"requiresApprovingReviews": true,
			},
		},
		{
			name: "branchProtectionQuery returns create query",
			args: args{
				branchProtectionArgs: []BranchProtectionArgs{
					{
						Name:     "requiresApprovingReviews",
						DataType: "Boolean",
						Value:    true,
					},
					{
						Name:     "repositoryId",
						DataType: "String",
						Value:    "some-repo-id",
					},
				},
				action: "create",
			},
			filePath: "../testdata/mockCreateBranchProtectionQuery.txt",
			wantRequestVars: map[string]interface{}{
				"repositoryId":             "some-repo-id",
				"requiresApprovingReviews": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReturn, err := ioutil.ReadFile(tt.filePath)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}
			want := string(mockReturn)

			gotQuery, gotRequestVars := branchProtectionQuery(tt.args.branchProtectionArgs, tt.args.action)
			if gotQuery != want {
				t.Errorf("branchProtectionQuery() gotQuery = \n\n%v, want \n\n%v", gotQuery, want)
			}
			if !reflect.DeepEqual(gotRequestVars, tt.wantRequestVars) {
				t.Errorf("branchProtectionQuery() gotRequestVars = %v, want %v", gotRequestVars, tt.wantRequestVars)
			}
		})
	}
}

func Test_branchProtectionRequest(t *testing.T) {
	type args struct {
		query       string
		requestVars map[string]interface{}
	}

	tests := []struct {
		name      string
		args      args
		wantQuery string
		wantVars  map[string]interface{}
	}{
		{
			name: "branchProtectionRequest check one request var",
			args: args{
				query:       "some query",
				requestVars: map[string]interface{}{"request_var1": "requestvar1value"},
			},
			wantQuery: "some query",
			wantVars: map[string]interface{}{
				"clientMutationId": "github-admin-tool",
				"request_var1":     "requestvar1value",
			},
		},
		{
			name: "branchProtectionRequest check two request vars",
			args: args{
				query: "some query",
				requestVars: map[string]interface{}{
					"request_var1": "requestvar1value",
					"request_var2": "requestvar2value",
				},
			},
			wantQuery: "some query",
			wantVars: map[string]interface{}{
				"clientMutationId": "github-admin-tool",
				"request_var1":     "requestvar1value",
				"request_var2":     "requestvar2value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := branchProtectionRequest(tt.args.query, tt.args.requestVars)

			if !reflect.DeepEqual(got.Query(), tt.wantQuery) {
				t.Errorf("branchProtectionRequest() = %v, want %v", got.Query(), tt.wantQuery)
			}

			if !reflect.DeepEqual(got.Vars(), tt.wantVars) {
				t.Errorf("branchProtectionRequest() = %v, want %v", got.Vars(), tt.wantVars)
			}
		})
	}
}

func Test_branchProtectionUpdate(t *testing.T) {
	type args struct {
		branchProtectionArgs   []BranchProtectionArgs
		branchProtectionRuleID string
	}

	doBranchProtectionSend = mockDoBranchProtectionSend
	defer func() { doBranchProtectionSend = branchProtectionSend }()

	tests := []struct {
		name               string
		args               args
		wantErr            bool
		mockErrorFunctions bool
	}{
		{
			name: "branchProtectionUpdate is successful",
			args: args{
				branchProtectionRuleID: "some-rule-id",
			},
			wantErr:            false,
			mockErrorFunctions: false,
		},
		{
			name: "branchProtectionUpdate is successful",
			args: args{
				branchProtectionRuleID: "some-rule-id",
			},
			wantErr:            true,
			mockErrorFunctions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErrorFunctions {
				doBranchProtectionSend = mockDoBranchProtectionSendError
				defer func() { doBranchProtectionSend = branchProtectionSend }()
			}
			if err := branchProtectionUpdate(
				tt.args.branchProtectionArgs,
				tt.args.branchProtectionRuleID,
			); (err != nil) != tt.wantErr {
				t.Errorf("branchProtectionUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_branchProtectionCreate(t *testing.T) {
	type args struct {
		branchProtectionArgs []BranchProtectionArgs
		repositoryID         string
		pattern              string
	}

	doBranchProtectionSend = mockDoBranchProtectionSend
	defer func() { doBranchProtectionSend = branchProtectionSend }()

	tests := []struct {
		name               string
		args               args
		wantErr            bool
		mockErrorFunctions bool
	}{
		{
			name: "branchProtectionCreate is successful",
			args: args{
				repositoryID: "some-repo-id",
				pattern:      "branch-name",
			},
			wantErr:            false,
			mockErrorFunctions: false,
		},
		{
			name: "branchProtectionCreate is successful",
			args: args{
				repositoryID: "some-repo-id",
				pattern:      "branch-name",
			},
			wantErr:            true,
			mockErrorFunctions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErrorFunctions {
				doBranchProtectionSend = mockDoBranchProtectionSendError
				defer func() { doBranchProtectionSend = branchProtectionSend }()
			}
			if err := branchProtectionCreate(
				tt.args.branchProtectionArgs,
				tt.args.repositoryID,
				tt.args.pattern,
			); (err != nil) != tt.wantErr {
				t.Errorf("branchProtectionCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_branchProtectionSend(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		req    *graphqlclient.Request
		client *graphqlclient.Client
	}

	tests := []struct {
		name               string
		args               args
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "branchProtectionSend success",
			args: args{
				req:    graphqlclient.NewRequest("query"),
				client: graphqlclient.NewClient("https://api.github.com/graphql"),
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "branchProtectionSend success",
			args: args{
				req:    graphqlclient.NewRequest("query"),
				client: graphqlclient.NewClient("https://api.github.com/graphql"),
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateErrorJsonResponse.json",
			mockHTTPStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPReturn, err := ioutil.ReadFile(tt.mockHTTPReturnFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			httpmock.RegisterResponder(
				"POST",
				"https://api.github.com/graphql",
				httpmock.NewStringResponder(tt.mockHTTPStatusCode, string(mockHTTPReturn)),
			)

			if err := branchProtectionSend(tt.args.req, tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("branchProtectionSend() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
