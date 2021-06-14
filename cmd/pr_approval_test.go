package cmd

import (
	"github-admin-tool/graphqlclient"
	"reflect"
	"testing"
)

func Test_applyPrApproval(t *testing.T) {
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
						DefaultBranchRef: DefaultBranchRef {
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



	}

	

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
