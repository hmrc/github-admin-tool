package cmd

import (
	"errors"
	"fmt"
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func Test_readRepoList(t *testing.T) {
	type args struct {
		reposFile string
	}

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "ReadRepoListReturnsError",
			args: args{
				reposFile: "blah.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ReadRepoListReturnsOneRepo",
			args: args{
				reposFile: "../testdata/one_repo_list.txt",
			},
			want:    []string{"a-test-repo"},
			wantErr: false,
		},
		{
			name: "ReadRepoListReturnsTwoRepo",
			args: args{
				reposFile: "../testdata/two_repo_list.txt",
			},
			want:    []string{"a-test-repo", "a-test-repo2"},
			wantErr: false,
		},
		{
			name: "ReadRepoListThrowsError",
			args: args{
				reposFile: "../testdata/repo_list_bad.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readRepoList(tt.args.reposFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readRepoList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readRepoList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateRepoQuery(t *testing.T) {
	type args struct {
		repos []string
	}

	tests := []struct {
		name           string
		args           args
		wantsToContain []string
	}{
		{
			name: "GenerateRepoQueryWithOneRepo",
			args: args{
				repos: []string{"repo-name-1"},
			},
			wantsToContain: []string{`repo0: repository(owner: $org, name: "repo-name-1")`},
		},
		{
			name: "GenerateRepoQueryWithTwoRepos",
			args: args{
				repos: []string{"repo-name-1", "repo-name-2"},
			},
			wantsToContain: []string{
				`repo0: repository(owner: $org, name: "repo-name-1")`,
				`repo1: repository(owner: $org, name: "repo-name-2")`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateRepoQuery(tt.args.repos)
			for _, wantsToContain := range tt.wantsToContain {
				if !strings.Contains(got, wantsToContain) {
					t.Errorf("generateRepoQuery() = %v, wantsToContains %v'", got, wantsToContain)
				}
			}
		})
	}
}

func Test_repoRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	type args struct {
		queryString string
		client      *graphqlclient.Client
	}

	tests := []struct {
		name               string
		args               args
		want               map[string]RepositoriesNodeList
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "GetRepoResponse",
			args: args{
				queryString: "", client: client,
			},
			want: map[string]RepositoriesNodeList{"repo0": {
				ID:            "repoIdTEST",
				NameWithOwner: "org/some-repo-name",
			}},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockRepoJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "GetRepoResponseWithError",
			args: args{
				queryString: "", client: client,
			},
			want:               map[string]RepositoriesNodeList{"repo0": {}},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockEmptyBranchProtectionJsonResponse.json",
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

			got, err := repoRequest(tt.args.queryString, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("repoRequest() error = %+v, wantErr %+v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("repoRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createQueryBlocks(t *testing.T) {
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
			name: "createQueryBlocks",
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
			name: "createQueryBlocksWithMoreThanOneArg",
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
			gotMutation, gotInput, gotRequestVars := createQueryBlocks(tt.args.args)
			if !reflect.DeepEqual(gotMutation.String(), tt.wantMutation) {
				t.Errorf("createQueryBlocks() gotMutation = %v, want %v", gotMutation.String(), tt.wantMutation)
			}
			if !reflect.DeepEqual(gotInput.String(), tt.wantInput) {
				t.Errorf("createQueryBlocks() gotInput = %v, want = %v", gotInput.String(), tt.wantInput)
			}
			if !reflect.DeepEqual(gotRequestVars, tt.wantRequestVars) {
				t.Errorf("createQueryBlocks() gotRequestVars = %v, want %v", gotRequestVars, tt.wantRequestVars)
			}
		})
	}
}

func Test_updateBranchProtection(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	type args struct {
		branchProtectionRuleID string
		branchProtectionArgs   []BranchProtectionArgs
		client                 *graphqlclient.Client
	}

	tests := []struct {
		name               string
		args               args
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "updateBranchProtectionSuccess",
			args: args{
				branchProtectionRuleID: "some-id",
				branchProtectionArgs: []BranchProtectionArgs{{
					Name:     "requiresApprovingReviews",
					DataType: "Boolean",
					Value:    "true",
				}},
				client: client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "updateBranchProtectionError",
			args: args{
				branchProtectionRuleID: "some-id",
				branchProtectionArgs: []BranchProtectionArgs{{
					Name:     "requiresApprovingReviews",
					DataType: "Boolean",
					Value:    "true",
				}},
				client: client,
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

			if err := updateBranchProtection(
				tt.args.branchProtectionRuleID,
				tt.args.branchProtectionArgs,
				tt.args.client,
			); (err != nil) != tt.wantErr {
				t.Errorf("updateBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createBranchProtection(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	type args struct {
		repositoryID         string
		branchName           string
		branchProtectionArgs []BranchProtectionArgs
		client               *graphqlclient.Client
	}

	tests := []struct {
		name               string
		args               args
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "createBranchProtectionSuccess",
			args: args{
				repositoryID: "some-repo-id",
				branchName:   "some-branch-name",
				branchProtectionArgs: []BranchProtectionArgs{{
					Name:     "requiresApprovingReviews",
					DataType: "Boolean",
					Value:    "true",
				}},
				client: client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionCreateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "createBranchProtectionError",
			args: args{
				repositoryID: "some-repo-id",
				branchName:   "some-branch-name",
				branchProtectionArgs: []BranchProtectionArgs{{
					Name:     "requiresApprovingReviews",
					DataType: "Boolean",
					Value:    "true",
				}},
				client: client,
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionCreateErrorJsonResponse.json",
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

			if err := createBranchProtection(
				tt.args.repositoryID,
				tt.args.branchName,
				tt.args.branchProtectionArgs,
				tt.args.client,
			); (err != nil) != tt.wantErr {
				t.Errorf("updateBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var errMockTest = errors.New("test")

func mockedUpdateBranchProtection(bpID string, args []BranchProtectionArgs, client *graphqlclient.Client) error {
	return nil
}

func mockedCreateBranchProtection(
	repositoryID,
	branchName string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return nil
}

func mockedUpdateBranchProtectionError(
	branchProtectionID string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return fmt.Errorf("update: %w", errMockTest)
}

func mockedCreateBranchProtectionError(
	repositoryID,
	branchName string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return fmt.Errorf("create: %w", errMockTest)
}

func Test_applyBranchProtection(t *testing.T) {
	type args struct {
		repoSearchResult     map[string]RepositoriesNodeList
		action               string
		branchProtectionArgs []BranchProtectionArgs
		client               *graphqlclient.Client
	}

	updateBranchProtectionRule = mockedUpdateBranchProtection
	defer func() { updateBranchProtectionRule = updateBranchProtection }()

	createBranchProtectionRule = mockedCreateBranchProtection
	defer func() { createBranchProtectionRule = createBranchProtection }()

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
			name: "ApplyBranchProtectionWithNoDefaultBranch",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{"repo0": {
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
			name: "ApplyBranchProtectionWithNoDefaultBranchBPR",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/no-branch-protection",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNodesList{{
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
			name: "ApplyBranchProtectionWithDefaultBranchBPRSigningOn",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/signing-on",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNodesList{{
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
			name: "ApplyBranchProtectionWithDefaultBranchBPRSigningOff",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/signing-off",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNodesList{{
							RequiresCommitSignatures: false,
							Pattern:                  "default-branch-name",
						}},
					},
				}},
			},
			wantModified: []string{"org/signing-off"},
			wantCreated:  nil,
			wantInfo:     nil,
			wantErrors:   nil,
		},
		{
			name: "ApplySigningCreatingFailure",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{"repo0": {
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
			name: "ApplySigningUpdatingFailure",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{"repo0": {
					ID:            "repoIdTEST",
					NameWithOwner: "org/signing-off",
					DefaultBranchRef: DefaultBranchRef{
						Name: "default-branch-name",
					},
					BranchProtectionRules: BranchProtectionRules{
						Nodes: []BranchProtectionRulesNodesList{{
							RequiresCommitSignatures: false,
							Pattern:                  "default-branch-name",
						}},
					},
				}},
			},
			wantModified:       nil,
			wantCreated:        nil,
			wantInfo:           nil,
			wantErrors:         []string{"update: test"},
			mockErrorFunctions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockErrorFunctions {
				updateBranchProtectionRule = mockedUpdateBranchProtectionError
				defer func() { updateBranchProtectionRule = updateBranchProtection }()

				createBranchProtectionRule = mockedCreateBranchProtectionError
				defer func() { createBranchProtectionRule = createBranchProtection }()
			}

			gotModified, gotCreated, gotInfo, gotErrors := applyBranchProtection(
				tt.args.repoSearchResult,
				tt.args.action,
				tt.args.branchProtectionArgs,
				tt.args.client,
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
