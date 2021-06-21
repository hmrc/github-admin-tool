package cmd

import (
	"github-admin-tool/graphqlclient"
	"reflect"
	"testing"
)

func mockedUpdateBranchProtection(
	bprid string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return nil
}

func mockedCreateBranchProtection(
	rid,
	branchName string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return nil
}

func mockedUpdateBranchProtectionError(
	bprid string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return errors.New("Test update error")
}

func mockedCreateBranchProtectionError(
	rid,
	branchName string,
	args []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	return errors.New("Test create error")
}

func Test_applyPrApproval(t *testing.T) {
	prApprovalUpdate = mockedUpdateBranchProtection
	defer func() { prApprovalUpdate = updateBranchProtection }()

	prApprovalCreate = mockedCreateBranchProtection
	defer func() { prApprovalCreate = createBranchProtection }()

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
				prApprovalUpdate = mockedUpdateBranchProtectionError
				prApprovalCreate = mockedCreateBranchProtectionError
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