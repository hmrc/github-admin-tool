package cmd

import (
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
