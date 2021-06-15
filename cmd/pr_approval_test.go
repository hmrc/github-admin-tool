package cmd

import (
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
)

func mockedUpdatePrApprovalBranchProtection(branchProtectionRuleID string, client *graphqlclient.Client) error {
	return nil
}

func mockedCreatePrApprovalBranchProtection(repositoryID, branchName string, client *graphqlclient.Client) error {
	return nil
}

func mockedUpdatePrApprovalBranchProtectionError(branchProtectionRuleID string, client *graphqlclient.Client) error {
	return errors.New("Test update error")
}

func mockedCreatePrApprovalBranchProtectionError(repositoryID, branchName string, client *graphqlclient.Client) error {
	return errors.New("Test create error")
}

func Test_applyPrApproval(t *testing.T) {
	prApprovalUpdate = mockedUpdatePrApprovalBranchProtection
	defer func() { prApprovalUpdate = updatePrApprovalBranchProtection }()

	prApprovalCreate = mockedCreatePrApprovalBranchProtection
	defer func() { prApprovalCreate = createPrApprovalBranchProtection }()

	type args struct {
		repoSearchResult map[string]RepositoriesNodeList
		client           *graphqlclient.Client
	}

	tests := []struct {
		name         string
		args         args
		wantModified []string
		wantCreated  []string
		wantInfo     []string
		wantProblems []string
		returnError  bool
	}{
		{
			name: "applyPrApprovalNoDefaultBranch",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{
					"repo0": {
						ID:            "repoIdTEST",
						NameWithOwner: "org/some-repo-name",
					},
				},
			},
			wantInfo: []string{"No default branch for org/some-repo-name"},
		},
		{
			name: "applyPrApprovalUpdate",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{
					"repo0": {
						ID:            "repoIdTEST",
						NameWithOwner: "org/some-repo-name",
						DefaultBranchRef: DefaultBranchRef{
							Name: "some-branch-name",
						},
						BranchProtectionRules: BranchProtectionRules{
							Nodes: []BranchProtectionRulesNodesList{{
								Pattern: "some-branch-name",
							}},
						},
					},
				},
			},
			wantModified: []string{"org/some-repo-name"},
		},
		{
			name: "applyPrApprovalCreate",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{
					"repo0": {
						ID:            "repoIdTEST",
						NameWithOwner: "org/some-repo-name",
						DefaultBranchRef: DefaultBranchRef{
							Name: "some-branch-name",
						},
					},
				},
			},
			wantCreated: []string{"org/some-repo-name"},
		},
		{
			name: "applyPrApprovalUpdateError",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{
					"repo0": {
						ID:            "repoIdTEST",
						NameWithOwner: "org/some-repo-name",
						DefaultBranchRef: DefaultBranchRef{
							Name: "some-branch-name",
						},
						BranchProtectionRules: BranchProtectionRules{
							Nodes: []BranchProtectionRulesNodesList{{
								Pattern: "some-branch-name",
							}},
						},
					},
				},
			},
			wantProblems: []string{"Test update error"},
			returnError:  true,
		},
		{
			name: "applyPrApprovalCreateError",
			args: args{
				repoSearchResult: map[string]RepositoriesNodeList{
					"repo0": {
						ID:            "repoIdTEST",
						NameWithOwner: "org/some-repo-name",
						DefaultBranchRef: DefaultBranchRef{
							Name: "some-branch-name",
						},
					},
				},
			},
			wantProblems: []string{"Test create error"},
			returnError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.returnError {
				prApprovalUpdate = mockedUpdatePrApprovalBranchProtectionError
				prApprovalCreate = mockedCreatePrApprovalBranchProtectionError
			}

			gotModified, gotCreated, gotInfo, gotProblems := applyPrApproval(tt.args.repoSearchResult, tt.args.client)
			if !reflect.DeepEqual(gotModified, tt.wantModified) {
				t.Errorf("applyPrApproval() gotModified = %v, want %v", gotModified, tt.wantModified)
			}
			if !reflect.DeepEqual(gotCreated, tt.wantCreated) {
				t.Errorf("applyPrApproval() gotCreated = %v, want %v", gotCreated, tt.wantCreated)
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("applyPrApproval() gotInfo = %v, want %v", gotInfo, tt.wantInfo)
			}
			if !reflect.DeepEqual(gotProblems, tt.wantProblems) {
				t.Errorf("applyPrApproval() gotProblems = %v, want %v", gotProblems, tt.wantProblems)
			}
		})
	}
}

func Test_updatePrApprovalBranchProtection(t *testing.T) {
	type args struct {
		branchProtectionRuleID string
		client                 *graphqlclient.Client
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	tests := []struct {
		name               string
		args               args
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "updatePrApprovalBranchProtectionisSuccessful",
			args: args{
				branchProtectionRuleID: "some-branch-protection-rule-id",
				client:                 client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},

		{
			name: "updatePrApprovalBranchProtectionError",
			args: args{
				branchProtectionRuleID: "some-branch-protection-rule-id",
				client:                 client,
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionUpdateErrorJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "updatePrApprovalBranchProtectionError400",
			args: args{
				branchProtectionRuleID: "some-branch-protection-rule-id",
				client:                 client,
			},
			wantErr:            true,
			mockHTTPStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockHTTPReturn []byte
			var err error
			if tt.mockHTTPReturnFile != "" {
				mockHTTPReturn, err = ioutil.ReadFile(tt.mockHTTPReturnFile)
				if err != nil {
					t.Fatalf("failed to read test data: %v", err)
				}
			}

			httpmock.RegisterResponder(
				"POST",
				"https://api.github.com/graphql",
				httpmock.NewStringResponder(tt.mockHTTPStatusCode, string(mockHTTPReturn)),
			)
			err = updatePrApprovalBranchProtection(tt.args.branchProtectionRuleID, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("updatePrApprovalBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createPrApprovalBranchProtection(t *testing.T) {
	type args struct {
		repositoryID           string
		branchProtectionRuleID string
		branchName             string
		client                 *graphqlclient.Client
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	tests := []struct {
		name               string
		args               args
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "CreatePrApprovalBranchProtectionSuccess",
			args: args{
				repositoryID: "some-repo-id",
				client:       client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionCreateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "CreatePrApprovalBranchProtectionError",
			args: args{
				repositoryID: "some-repo-id",
				client:       client,
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockBranchProtectionCreateErrorJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "CreatePrapprovalBranchProtectionError400",
			args: args{
				branchProtectionRuleID: "some-branch-protection-rule-id",
				client:                 client,
			},
			wantErr:            true,
			mockHTTPStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockHTTPReturn []byte
			var err error
			if tt.mockHTTPReturnFile != "" {
				mockHTTPReturn, err = ioutil.ReadFile(tt.mockHTTPReturnFile)
				if err != nil {
					t.Fatalf("failed to read test data: %v", err)
				}
			}

			httpmock.RegisterResponder(
				"POST",
				"https://api.github.com/graphql",
				httpmock.NewStringResponder(tt.mockHTTPStatusCode, string(mockHTTPReturn)),
			)

			err = createPrApprovalBranchProtection(tt.args.repositoryID, tt.args.branchName, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPrApprovalBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
