package cmd

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github-admin-tool/graphqlclient"
	"github.com/jarcoal/httpmock"
)

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
			mockHTTPReturnFile: "../testdata/mockSigningJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "GetRepoResponseWithError",
			args: args{
				queryString: "", client: client,
			},
			want:               map[string]RepositoriesNodeList{"repo0": {}},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockEmptySigningJsonResponse.json",
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

func mockedUpdateBranchProtection(branchProtectionID string, client *graphqlclient.Client) error {
	return nil
}

func mockedCreateBranchProtection(repositoryID, branchName string, client *graphqlclient.Client) error {
	return nil
}

func Test_applySigning(t *testing.T) {
	type args struct {
		repoSearchResult map[string]RepositoriesNodeList
		client           *graphqlclient.Client
	}

	signingUpdateFunction = mockedUpdateBranchProtection
	defer func() { signingUpdateFunction = updateBranchProtection }()

	signingCreateFunction = mockedCreateBranchProtection
	defer func() { signingCreateFunction = createBranchProtection }()

	tests := []struct {
		name         string
		args         args
		wantModified []string
		wantInfo     []string
		wantErrors   []string
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
				}},
			},
			wantModified: []string{"org/no-branch-protection"},
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
			wantInfo:     nil,
			wantErrors:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModified, gotInfo, gotErrors := applySigning(tt.args.repoSearchResult, tt.args.client)
			if !reflect.DeepEqual(gotModified, tt.wantModified) {
				t.Errorf("applySigning() gotModified = %v, want %v", gotModified, tt.wantModified)
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

func Test_updateBranchProtection(t *testing.T) {
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
			name: "UpdateBranchProtectionSuccess",
			args: args{
				branchProtectionID: "some-random-bpr-id",
				client:             client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockSigningUpdateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "UpdateBranchProtectionFailure",
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

			if err := updateBranchProtection(
				tt.args.branchProtectionID,
				tt.args.client,
			); (err != nil) != tt.wantErr {
				t.Errorf("updateBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_createBranchProtection(t *testing.T) {
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
			name: "CreeateBranchProtectionSuccess",
			args: args{
				repositoryID: "some-repo-id",
				client:       client,
			},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockSigningCreateJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "CreeateBranchProtectionSuccess",
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

			if err := createBranchProtection(
				tt.args.repositoryID,
				tt.args.branchName,
				tt.args.client,
			); (err != nil) != tt.wantErr {
				t.Errorf("createBranchProtection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateQuery(t *testing.T) {
	type args struct {
		repos []string
	}

	tests := []struct {
		name           string
		args           args
		wantsToContain []string
	}{
		{
			name: "GenerateQueryWithOneRepo",
			args: args{
				repos: []string{"repo-name-1"},
			},
			wantsToContain: []string{`repo0: repository(owner: $org, name: "repo-name-1")`},
		},
		{
			name: "GenerateQueryWithTwoRepos",
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
			got := generateQuery(tt.args.repos)
			for _, wantsToContain := range tt.wantsToContain {
				if !strings.Contains(got, wantsToContain) {
					t.Errorf("generateQuery() = %v, wantsToContains %v'", got, wantsToContain)
				}
			}
		})
	}
}

func Test_readList(t *testing.T) {
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
			name: "ReadListReturnsError",
			args: args{
				reposFile: "blah.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ReadListReturnsOneRepo",
			args: args{
				reposFile: "../testdata/one_repo_list.txt",
			},
			want:    []string{"a-test-repo"},
			wantErr: false,
		},
		{
			name: "ReadListReturnsTwoRepo",
			args: args{
				reposFile: "../testdata/two_repo_list.txt",
			},
			want:    []string{"a-test-repo", "a-test-repo2"},
			wantErr: false,
		},
		{
			name: "ReadListThrowsError",
			args: args{
				reposFile: "../testdata/repo_list_bad.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readList(tt.args.reposFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("readList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readList() = %v, want %v", got, tt.want)
			}
		})
	}
}
