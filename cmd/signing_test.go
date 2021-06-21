package cmd

import (
	"errors"
	"fmt"
	"github-admin-tool/graphqlclient"
	"reflect"
	"testing"
)

var errMockTest = errors.New("test")

func mockedUpdateSigningBranchProtection(bpID string, args []BranchProtectionArgs, client *graphqlclient.Client) error {
	return nil
}

func mockedCreateSigningBranchProtection(
	repositoryID,
	branchName string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return nil
}

func mockedUpdateSigningBranchProtectionError(
	branchProtectionID string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return fmt.Errorf("update: %w", errMockTest)
}

func mockedCreateSigningBranchProtectionError(
	repositoryID,
	branchName string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return fmt.Errorf("create: %w", errMockTest)
}

func Test_applySigning(t *testing.T) {
	type args struct {
		repoSearchResult map[string]RepositoriesNodeList
		client           *graphqlclient.Client
	}

	signingUpdate = mockedUpdateSigningBranchProtection
	defer func() { signingUpdate = updateBranchProtection }()

	signingCreate = mockedCreateSigningBranchProtection
	defer func() { signingCreate = createBranchProtection }()

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
				defer func() { signingUpdate = updateBranchProtection }()

				signingCreate = mockedCreateSigningBranchProtectionError
				defer func() { signingCreate = createBranchProtection }()
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
