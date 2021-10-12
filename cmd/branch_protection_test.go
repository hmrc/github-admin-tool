package cmd

import (
	"errors"
	"fmt"
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/cobra"
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

type testSender struct {
	sendFail bool
	action   string
}

func (t *testSender) send(req *graphqlclient.Request) error {
	if t.sendFail {
		return errors.New(fmt.Sprintf("%s: test", t.action)) // nolint // only mock error for test
	}

	return nil
}

func Test_branchProtectionApply(t *testing.T) {
	type args struct {
		repoSearchResult     map[string]*RepositoriesNode
		action               string
		branchName           string
		branchProtectionArgs []BranchProtectionArgs
		sender               *githubBranchProtectionSender
	}

	tests := []struct {
		name         string
		args         args
		wantModified []string
		wantCreated  []string
		wantInfo     []string
		wantErrors   []string
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
			name: "branchProtectionApply signing with no default branch protection rule",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/no-branch-protection",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
				}},
				action: "Signing",
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
			wantModified: nil,
			wantCreated: []string{
				"Branch protection rule created for org/no-branch-protection with branch name: default-branch-name",
			},
			wantInfo:   nil,
			wantErrors: nil,
		},
		{
			name: "branchProtectionApply signing with no default branch protection rule and additional rule",
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
				action: "Signing",
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
			wantModified: nil,
			wantCreated: []string{
				"Branch protection rule created for org/no-branch-protection with branch name: default-branch-name",
			},
			wantInfo: []string{
				"Signing already turned on for org/no-branch-protection with branch name: another-branch-name",
			},
			wantErrors: nil,
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
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
			wantModified: nil,
			wantCreated:  nil,
			wantInfo:     []string{"Signing already turned on for org/signing-on with branch name: default-branch-name"},
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
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
			wantModified: []string{"Signing changed for org/signing-off with branch name: default-branch-name"},
			wantCreated:  nil,
			wantInfo:     nil,
			wantErrors:   nil,
		},
		{
			name: "branchProtectionApply signing with multiple rules",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/signing-off",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNode{
							{
								RequiresCommitSignatures: false,
								Pattern:                  "default-branch-name",
							},
							{
								RequiresCommitSignatures: false,
								Pattern:                  "another-branch-name",
							},
						},
					},
				}},
				action: "Signing",
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
			wantModified: []string{
				"Signing changed for org/signing-off with branch name: default-branch-name",
				"Signing changed for org/signing-off with branch name: another-branch-name",
			},
			wantCreated: nil,
			wantInfo:    nil,
			wantErrors:  nil,
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
				sender: &githubBranchProtectionSender{
					sender: &testSender{
						sendFail: true,
						action:   "create",
					},
				},
			},
			wantModified: nil,
			wantCreated:  nil,
			wantInfo:     nil,
			wantErrors:   []string{"create: test"},
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
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
			wantModified: nil,
			wantCreated:  nil,
			wantInfo: []string{
				"Pr-approval already turned on for org/pr-approval-duplicate with branch name: default-branch-name",
			},
			wantErrors: nil,
		},
		{
			name: "branchProtectionApply with multiple branch protection rules",
			args: args{
				repoSearchResult: map[string]*RepositoriesNode{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/pr-approval-duplicate",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNode{
							{
								RequiresApprovingReviews:     true,
								RequiredApprovingReviewCount: 1,
								DismissesStaleReviews:        true,
								RequiresCodeOwnerReviews:     false,
								Pattern:                      "default-branch-name",
							},
							{
								RequiresApprovingReviews:     true,
								RequiredApprovingReviewCount: 1,
								DismissesStaleReviews:        true,
								RequiresCodeOwnerReviews:     false,
								Pattern:                      "main-name",
							},
						},
					},
				}},
				action:     "Pr-approval",
				branchName: "main",
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
			wantModified: nil,
			wantCreated: []string{
				"Branch protection rule created for org/pr-approval-duplicate with branch name: main",
			},
			wantInfo:   nil,
			wantErrors: nil,
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
				sender: &githubBranchProtectionSender{
					sender: &testSender{
						sendFail: true,
						action:   "update",
					},
				},
			},
			wantErrors: []string{"update: test"},
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
				sender: &githubBranchProtectionSender{
					sender: &testSender{
						sendFail: true,
						action:   "create",
					},
				},
			},
			wantErrors: []string{"create: test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModified, gotCreated, gotInfo, gotErrors := branchProtectionApply(
				tt.args.repoSearchResult,
				tt.args.action,
				tt.args.branchName,
				tt.args.branchProtectionArgs,
				tt.args.sender,
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
				t.Errorf("branchProtectionRequest() query = %v, want %v", got.Query(), tt.wantQuery)
			}

			if !reflect.DeepEqual(got.Vars(), tt.wantVars) {
				t.Errorf("branchProtectionRequest() vars = %v, want %v", got.Vars(), tt.wantVars)
			}
		})
	}
}

func Test_branchProtectionUpdate(t *testing.T) {
	type args struct {
		branchProtectionArgs   []BranchProtectionArgs
		branchProtectionRuleID string
		sender                 *githubBranchProtectionSender
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "branchProtectionUpdate is successful",
			args: args{
				branchProtectionRuleID: "some-rule-id",
				sender: &githubBranchProtectionSender{
					sender: &testSender{},
				},
			},
			wantErr: false,
		},
		{
			name: "branchProtectionUpdate is successful",
			args: args{
				branchProtectionRuleID: "some-rule-id",
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: true},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := branchProtectionUpdate(
				tt.args.branchProtectionArgs,
				tt.args.branchProtectionRuleID,
				tt.args.sender,
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
		sender               *githubBranchProtectionSender
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "branchProtectionCreate is successful",
			args: args{
				repositoryID: "some-repo-id",
				pattern:      "branch-name",
				sender: &githubBranchProtectionSender{
					sender: &testSender{},
				},
			},
			wantErr: false,
		},
		{
			name: "branchProtectionCreate is successful",
			args: args{
				repositoryID: "some-repo-id",
				pattern:      "branch-name",
				sender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: true},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := branchProtectionCreate(
				tt.args.branchProtectionArgs,
				tt.args.repositoryID,
				tt.args.pattern,
				tt.args.sender,
			); (err != nil) != tt.wantErr {
				t.Errorf("branchProtectionCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_branchProtectionSenderService_send(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		req *graphqlclient.Request
	}

	tests := []struct {
		name               string
		b                  *branchProtectionSenderService
		args               args
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "branchProtectionSend success",
			args: args{
				req: graphqlclient.NewRequest("query"),
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "branchProtectionSend success",
			args: args{
				req: graphqlclient.NewRequest("query"),
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateErrorJsonResponse.json",
			mockHTTPStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &branchProtectionSenderService{}
			mockHTTPReturn, err := ioutil.ReadFile(tt.mockHTTPReturnFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			httpmock.RegisterResponder(
				"POST",
				"https://api.github.com/graphql",
				httpmock.NewStringResponder(tt.mockHTTPStatusCode, string(mockHTTPReturn)),
			)

			if err := b.send(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("branchProtectionSenderService.send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type testRepositoryReader struct {
	readFail    bool
	returnValue []string
}

func (t *testRepositoryReader) read(reposFile string) ([]string, error) {
	if t.readFail {
		return t.returnValue, errors.New("fail") // nolint // only mock error for test
	}

	return t.returnValue, nil
}

type testRepositoryGetter struct {
	getFail     bool
	returnValue map[string]*RepositoriesNode
}

func (t *testRepositoryGetter) get(
	repositoryList []string,
	sender *githubRepositorySender,
) (
	map[string]*RepositoriesNode,
	error,
) {
	if t.getFail {
		return t.returnValue, errors.New("fail") // nolint // only mock error for test
	}

	return t.returnValue, nil
}

func Test_branchProtectionCommand(t *testing.T) {
	type args struct {
		cmd                    *cobra.Command
		branchProtectionArgs   []BranchProtectionArgs
		action                 string
		branchName             string
		repo                   *repository
		repoSender             *githubRepositorySender
		branchProtectionSender *githubBranchProtectionSender
	}

	var (
		mockRepos2File  string
		mockDryRunFalse bool
	)

	mockCmdWithDryRunOff := &cobra.Command{
		Use: "pr-approval",
	}
	mockCmdWithDryRunOff.Flags().BoolVarP(&mockDryRunFalse, "dry-run", "d", false, "dry run flag")
	mockCmdWithDryRunOff.Flags().StringVarP(
		&mockRepos2File,
		"repos",
		"r",
		"../testdata/two_repo_list.txt",
		"repos file",
	)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "branchProtectionCommand is success",
			args: args{
				cmd: mockCmdWithDryRunOff,
				repo: &repository{
					reader: &testRepositoryReader{
						returnValue: []string{
							"some-repo-name",
						},
					},
					getter: &testRepositoryGetter{
						returnValue: map[string]*RepositoriesNode{"repo0": {
							ID:            "repoIdTEST",
							NameWithOwner: "org/some-repo-name",
						}},
					},
				},
				repoSender: &githubRepositorySender{
					sender: &testRepositorySender{
						returnValue: map[string]*RepositoriesNode{"repo0": {
							ID:            "repoIdTEST",
							NameWithOwner: "org/some-repo-name",
						}},
					},
				},
				branchProtectionSender: &githubBranchProtectionSender{
					sender: &testSender{sendFail: false},
				},
			},
		},
		{
			name: "branchProtectionCommand is failure",
			args: args{
				cmd: mockCmdWithDryRunOff,
				repo: &repository{
					reader: &testRepositoryReader{
						returnValue: []string{
							"some-repo-name",
						},
					},
					getter: &testRepositoryGetter{
						getFail: true,
					},
				},
				repoSender: &githubRepositorySender{
					sender: &testRepositorySender{},
				},
				branchProtectionSender: &githubBranchProtectionSender{
					sender: &testSender{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := branchProtectionCommand(
				tt.args.cmd,
				tt.args.branchProtectionArgs,
				tt.args.action,
				tt.args.branchName,
				tt.args.repo,
				tt.args.repoSender,
				tt.args.branchProtectionSender,
			); (err != nil) != tt.wantErr {
				t.Errorf("branchProtectionCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_branchProtectionDisplayInfo(t *testing.T) {
	type args struct {
		updated   []string
		created   []string
		info      []string
		problems  []string
		batchInfo string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "branchProtectionDisplayInfo all logs",
			args: args{
				updated:   []string{"updated-repo-name"},
				created:   []string{"created-repo-name"},
				info:      []string{"some-info"},
				problems:  []string{"some-problems"},
				batchInfo: "1-4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branchProtectionDisplayInfo(tt.args.updated, tt.args.created, tt.args.info, tt.args.problems, tt.args.batchInfo)
		})
	}
}
