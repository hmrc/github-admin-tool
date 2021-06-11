package cmd

import (
	"errors"
	"fmt"
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

var errMockTest = errors.New("test")

func mockedUpdateSigningBranchProtection(branchProtectionID string, client *graphqlclient.Client) error {
	return nil
}

func mockedCreateSigningBranchProtection(repositoryID, branchName string, client *graphqlclient.Client) error {
	return nil
}

func mockedUpdateSigningBranchProtectionError(branchProtectionID string, client *graphqlclient.Client) error {
	return fmt.Errorf("update: %w", errMockTest)
}

func mockedCreateSigningBranchProtectionError(repositoryID, branchName string, client *graphqlclient.Client) error {
	return fmt.Errorf("create: %w", errMockTest)
}

func Test_applySigning(t *testing.T) {
	type args struct {
		repoSearchResult map[string]RepositoriesNodeList
		client           *graphqlclient.Client
	}

	signingUpdate = mockedUpdateSigningBranchProtection
	defer func() { signingUpdate = updateSigningBranchProtection }()

	signingCreate = mockedCreateSigningBranchProtection
	defer func() { signingCreate = createSigningBranchProtection }()

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
			name: "ApplySigningWithNoDefaultBranch",
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
			name: "ApplySigningWithNoDefaultBranchBPR",
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
			name: "ApplySigningWithDefaultBranchBPRSigningOn",
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
			},
			wantModified: nil,
			wantCreated:  nil,
			wantInfo:     []string{"Signing already turned on for org/signing-on"},
			wantErrors:   nil,
		},
		{
			name: "ApplySigningWithDefaultBranchBPRSigningOff",
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
				signingUpdate = mockedUpdateSigningBranchProtectionError
				defer func() { signingUpdate = updateSigningBranchProtection }()

				signingCreate = mockedCreateSigningBranchProtectionError
				defer func() { signingCreate = createSigningBranchProtection }()
			}

			gotModified, gotCreated, gotInfo, gotErrors := applySigning(tt.args.repoSearchResult, tt.args.client)
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

func Test_updateSigningBranchProtection(t *testing.T) {
	type args struct {
		branchProtectionID string
		client             *graphqlclient.Client
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
			name: "UpdateSigningBranchProtectionSuccess",
			args: args{
				branchProtectionID: "some-random-bpr-id",
				client:             client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockSigningUpdateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "UpdateSigningBranchProtectionFailure",
			args: args{
				branchProtectionID: "some-random-bpr-id",
				client:             client,
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockSigningUpdateErrorJsonResponse.json",
			mockHTTPStatusCode: 200,
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

			if err := updateSigningBranchProtection(
				tt.args.branchProtectionID,
				tt.args.client,
			); (err != nil) != tt.wantErr {
				t.Errorf("updateSigningBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createSigningBranchProtection(t *testing.T) {
	type args struct {
		repositoryID string
		branchName   string
		client       *graphqlclient.Client
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
			name: "CreeateSigningBranchProtectionSuccess",
			args: args{
				repositoryID: "some-repo-id",
				client:       client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockSigningCreateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "CreeateSigningBranchProtectionSuccess",
			args: args{
				repositoryID: "some-repo-id",
				client:       client,
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockSigningCreateErrorJsonResponse.json",
			mockHTTPStatusCode: 200,
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

			if err := createSigningBranchProtection(
				tt.args.repositoryID,
				tt.args.branchName,
				tt.args.client,
			); (err != nil) != tt.wantErr {
				t.Errorf("createSigningBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
