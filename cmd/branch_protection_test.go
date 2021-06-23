package cmd

import (
	"reflect"
	"testing"
	"io/ioutil"
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

// func Test_updateBranchProtectionQuery(t *testing.T) {
// 	httpmock.Activate()
// 	defer httpmock.DeactivateAndReset()

// 	client := graphqlclient.NewClient("https://api.github.com/graphql")

// 	type args struct {
// 		branchProtectionRuleID string
// 		branchProtectionArgs   []BranchProtectionArgs
// 		client                 *graphqlclient.Client
// 	}

// 	tests := []struct {
// 		name               string
// 		args               args
// 		wantErr            bool
// 		mockHTTPReturnFile string
// 		mockHTTPStatusCode int
// 	}{
// 		{
// 			name: "updateBranchProtection success",
// 			args: args{
// 				branchProtectionRuleID: "some-id",
// 				branchProtectionArgs: []BranchProtectionArgs{{
// 					Name:     "requiresApprovingReviews",
// 					DataType: "Boolean",
// 					Value:    "true",
// 				}},
// 				client: client,
// 			},
// 			wantErr:            false,
// 			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateJsonResponse.json",
// 			mockHTTPStatusCode: 200,
// 		},
// 		{
// 			name: "updateBranchProtection error",
// 			args: args{
// 				branchProtectionRuleID: "some-id",
// 				branchProtectionArgs: []BranchProtectionArgs{{
// 					Name:     "requiresApprovingReviews",
// 					DataType: "Boolean",
// 					Value:    "true",
// 				}},
// 				client: client,
// 			},
// 			wantErr:            true,
// 			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateErrorJsonResponse.json",
// 			mockHTTPStatusCode: 400,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockHTTPReturn, err := ioutil.ReadFile(tt.mockHTTPReturnFile)
// 			if err != nil {
// 				t.Fatalf("failed to read test data: %v", err)
// 			}

// 			httpmock.RegisterResponder(
// 				"POST",
// 				"https://api.github.com/graphql",
// 				httpmock.NewStringResponder(tt.mockHTTPStatusCode, string(mockHTTPReturn)),
// 			)

// 			if err := updateBranchProtection(
// 				tt.args.branchProtectionRuleID,
// 				tt.args.branchProtectionArgs,
// 				tt.args.client,
// 			); (err != nil) != tt.wantErr {
// 				t.Errorf("updateBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_createBranchProtection(t *testing.T) {
// 	httpmock.Activate()
// 	defer httpmock.DeactivateAndReset()

// 	client := graphqlclient.NewClient("https://api.github.com/graphql")

// 	type args struct {
// 		repositoryID         string
// 		branchName           string
// 		branchProtectionArgs []BranchProtectionArgs
// 		client               *graphqlclient.Client
// 	}

// 	tests := []struct {
// 		name               string
// 		args               args
// 		wantErr            bool
// 		mockHTTPReturnFile string
// 		mockHTTPStatusCode int
// 	}{
// 		{
// 			name: "createBranchProtection success",
// 			args: args{
// 				repositoryID: "some-repo-id",
// 				branchName:   "some-branch-name",
// 				branchProtectionArgs: []BranchProtectionArgs{{
// 					Name:     "requiresApprovingReviews",
// 					DataType: "Boolean",
// 					Value:    "true",
// 				}},
// 				client: client,
// 			},
// 			wantErr:            false,
// 			mockHTTPReturnFile: "../testdata/mockBranchProtectionCreateJsonResponse.json",
// 			mockHTTPStatusCode: 200,
// 		},
// 		{
// 			name: "createBranchProtection error",
// 			args: args{
// 				repositoryID: "some-repo-id",
// 				branchName:   "some-branch-name",
// 				branchProtectionArgs: []BranchProtectionArgs{{
// 					Name:     "requiresApprovingReviews",
// 					DataType: "Boolean",
// 					Value:    "true",
// 				}},
// 				client: client,
// 			},
// 			wantErr:            true,
// 			mockHTTPReturnFile: "../testdata/mockBranchProtectionCreateErrorJsonResponse.json",
// 			mockHTTPStatusCode: 400,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockHTTPReturn, err := ioutil.ReadFile(tt.mockHTTPReturnFile)
// 			if err != nil {
// 				t.Fatalf("failed to read test data: %v", err)
// 			}

// 			httpmock.RegisterResponder(
// 				"POST",
// 				"https://api.github.com/graphql",
// 				httpmock.NewStringResponder(tt.mockHTTPStatusCode, string(mockHTTPReturn)),
// 			)

// 			if err := createBranchProtection(
// 				tt.args.repositoryID,
// 				tt.args.branchName,
// 				tt.args.branchProtectionArgs,
// 				tt.args.client,
// 			); (err != nil) != tt.wantErr {
// 				t.Errorf("updateBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// var errMockTest = errors.New("test")

// func mockBranchProtectionSend(
// 	query string,
// 	requestVars map[string]interface{},
// 	client *graphqlclient.Client,
// ) error {
// 	return nil
// }

// func mockedUpdateBranchProtectionError(
// 	query string,
// 	requestVars map[string]interface{},
// 	client *graphqlclient.Client,
// ) error {
// 	return fmt.Errorf("update: %w", errMockTest)
// }

// func Test_applyBranchProtection(t *testing.T) {
// 	type args struct {
// 		repoSearchResult     map[string]RepositoriesNode
// 		action               string
// 		branchProtectionArgs []BranchProtectionArgs
// 		client               *graphqlclient.Client
// 	}

// 	doBranchProtectionSend = mockBranchProtectionSend
// 	defer func() { doBranchProtectionSend = branchProtectionSend }()

// 	tests := []struct {
// 		name               string
// 		args               args
// 		wantModified       []string
// 		wantCreated        []string
// 		wantInfo           []string
// 		wantErrors         []string
// 		mockErrorFunctions bool
// 	}{
// 		{
// 			name: "applyBranchProtection with no default branch",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/some-repo-name",
// 				}},
// 			},
// 			wantModified: nil,
// 			wantCreated:  nil,
// 			wantInfo:     []string{"No default branch for org/some-repo-name"},
// 			wantErrors:   nil,
// 		},
// 		{
// 			name: "applyBranchProtection with no default branch protection rule",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/no-branch-protection",
// 					DefaultBranchRef: DefaultBranchRef{
// 						Name: "default-branch-name",
// 					},
// 					BranchProtectionRules: BranchProtectionRules{
// 						Nodes: []BranchProtectionRulesNode{{
// 							RequiresCommitSignatures: true,
// 							Pattern:                  "another-branch-name",
// 						}},
// 					},
// 				}},
// 			},
// 			wantModified: nil,
// 			wantCreated:  []string{"org/no-branch-protection"},
// 			wantInfo:     nil,
// 			wantErrors:   nil,
// 		},
// 		{
// 			name: "applyBranchProtection with default branch protection rule signing on",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/signing-on",
// 					DefaultBranchRef: DefaultBranchRef{
// 						Name: "default-branch-name",
// 					},
// 					BranchProtectionRules: BranchProtectionRules{
// 						Nodes: []BranchProtectionRulesNode{{
// 							RequiresCommitSignatures: true,
// 							Pattern:                  "default-branch-name",
// 						}},
// 					},
// 				}},
// 				action: "Signing",
// 			},
// 			wantModified: nil,
// 			wantCreated:  nil,
// 			wantInfo:     []string{"Signing already turned on for org/signing-on"},
// 			wantErrors:   nil,
// 		},
// 		{
// 			name: "applyBranchProtection with default branch protection rule signing off",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/signing-off",
// 					DefaultBranchRef: DefaultBranchRef{
// 						Name: "default-branch-name",
// 					},
// 					BranchProtectionRules: BranchProtectionRules{
// 						Nodes: []BranchProtectionRulesNode{{
// 							RequiresCommitSignatures: false,
// 							Pattern:                  "default-branch-name",
// 						}},
// 					},
// 				}},
// 				action: "Signing",
// 			},
// 			wantModified: []string{"org/signing-off"},
// 			wantCreated:  nil,
// 			wantInfo:     nil,
// 			wantErrors:   nil,
// 		},
// 		{
// 			name: "applyBranchProtection creating failure",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/signing-off",
// 					DefaultBranchRef: DefaultBranchRef{
// 						Name: "default-branch-name",
// 					},
// 				}},
// 			},
// 			wantModified:       nil,
// 			wantCreated:        nil,
// 			wantInfo:           nil,
// 			wantErrors:         []string{"create: test"},
// 			mockErrorFunctions: true,
// 		},
// 		{
// 			name: "applyBranchProtection updating failure",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/signing-off",
// 					DefaultBranchRef: DefaultBranchRef{
// 						Name: "default-branch-name",
// 					},
// 					BranchProtectionRules: BranchProtectionRules{
// 						Nodes: []BranchProtectionRulesNode{{
// 							RequiresCommitSignatures: false,
// 							Pattern:                  "default-branch-name",
// 						}},
// 					},
// 				}},
// 			},
// 			wantErrors:         []string{"update: test"},
// 			mockErrorFunctions: true,
// 		},
// 		{
// 			name: "applyBranchProtection pr approval update failure",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/pr-approval-test",
// 					DefaultBranchRef: DefaultBranchRef{
// 						Name: "default-branch-name",
// 					},
// 					BranchProtectionRules: BranchProtectionRules{
// 						Nodes: []BranchProtectionRulesNode{{
// 							RequiresCommitSignatures: true,
// 							Pattern:                  "default-branch-name",
// 						}},
// 					},
// 				}},
// 				action: "Pr-approval",
// 			},
// 			wantErrors:         []string{"update: test"},
// 			mockErrorFunctions: true,
// 		},
// 		{
// 			name: "applyBranchProtection pr approval create failure",
// 			args: args{
// 				repoSearchResult: map[string]RepositoriesNode{"repo0": {
// 					ID:            "repoIdTEST",
// 					NameWithOwner: "org/pr-approval-test",
// 					DefaultBranchRef: DefaultBranchRef{
// 						Name: "default-branch-name",
// 					},
// 				}},
// 				action: "Pr-approval",
// 			},
// 			wantErrors:         []string{"create: test"},
// 			mockErrorFunctions: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.mockErrorFunctions {
// 				doBranchProtectionRequest = mockedUpdateBranchProtectionError
// 				defer func() { doBranchProtectionRequest = branchProtectionRequest }()

// 				createBranchProtectionRule = mockedCreateBranchProtectionError
// 				defer func() { createBranchProtectionRule = createBranchProtection }()
// 			}

// 			gotModified, gotCreated, gotInfo, gotErrors := applyBranchProtection(
// 				tt.args.repoSearchResult,
// 				tt.args.action,
// 				tt.args.branchProtectionArgs,
// 				tt.args.client,
// 			)
// 			if !reflect.DeepEqual(gotModified, tt.wantModified) {
// 				t.Errorf("applySigning() gotModified = %v, want %v", gotModified, tt.wantModified)
// 			}
// 			if !reflect.DeepEqual(gotCreated, tt.wantCreated) {
// 				t.Errorf("applySigning() gotCreated = %v, want %v", gotCreated, tt.wantCreated)
// 			}
// 			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
// 				t.Errorf("applySigning() gotInfo = %v, want %v", gotInfo, tt.wantInfo)
// 			}
// 			if !reflect.DeepEqual(gotErrors, tt.wantErrors) {
// 				t.Errorf("applySigning() gotErrors = %v, want %v", gotErrors, tt.wantErrors)
// 			}
// 		})
// 	}
// }

// func Test_updateBranchProtectionQuery(t *testing.T) {
// 	type args struct {
// 		mutationBlock string
// 		inputBlock    string
// 	}

// 	tests := []struct {
// 		name     string
// 		args     args
// 		filePath string
// 	}{
// 		{
// 			name: "updateBranchProtectionQuery ",
// 			args: args{
// 				mutationBlock: "$requiresApprovingReviews: requiresApprovingReviews!,",
// 				inputBlock:    "requiresApprovingReviews: $requiresApprovingReviews,",
// 			},
// 			filePath: "../testdata/mockUpdateBranchProtectionQuery.txt",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockReturn, err := ioutil.ReadFile(tt.filePath)
// 			if err != nil {
// 				t.Fatalf("failed to read test data: %v", err)
// 			}

// 			want := string(mockReturn)

// 			if got := updateBranchProtectionQuery(tt.args.mutationBlock, tt.args.inputBlock); got != want {
// 				t.Errorf("updateBranchProtectionQuery() = \n\n%v, want \n\n%v", got, want)
// 			}
// 		})
// 	}
// }

func Test_branchProtectionQuery(t *testing.T) {
	type args struct {
		branchProtectionArgs []BranchProtectionArgs
		action               string
	}

	tests := []struct {
		name            string
		args            args
		filePath       string
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
				"branchProtectionRuleId": "some-rule-id",
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
				"repositoryId": "some-repo-id",
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
