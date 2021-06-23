package cmd

import (
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func Test_repoList(t *testing.T) {
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
			name: "repoList returns rrror",
			args: args{
				reposFile: "blah.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "repoList returns one repo",
			args: args{
				reposFile: "../testdata/one_repo_list.txt",
			},
			want:    []string{"a-test-repo"},
			wantErr: false,
		},
		{
			name: "repoList returns two repos",
			args: args{
				reposFile: "../testdata/two_repo_list.txt",
			},
			want:    []string{"a-test-repo", "a-test-repo2"},
			wantErr: false,
		},
		{
			name: "repoList throws error",
			args: args{
				reposFile: "../testdata/repo_list_bad.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repoList(tt.args.reposFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("repoList() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("repoList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_repoQuery(t *testing.T) {
	type args struct {
		repos []string
	}

	tests := []struct {
		name     string
		args     args
		filePath string
	}{
		{
			name: "repoQuery with one repo",
			args: args{
				repos: []string{"repo-name-1"},
			},
			filePath: "../testdata/mockGenerateRepoWithOneRepo.txt",
		},
		{
			name: "repoQuery with two repos",
			args: args{
				repos: []string{"repo-name-1", "repo-name-2"},
			},
			filePath: "../testdata/mockGenerateRepoWithTwoRepos.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReturn, err := ioutil.ReadFile(tt.filePath)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			want := string(mockReturn)

			got := repoQuery(tt.args.repos)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("repoQuery() = %s, mockReturn %s'", got, want)
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
		want               map[string]RepositoriesNode
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "repoRequest",
			args: args{
				queryString: "", client: client,
			},
			want: map[string]RepositoriesNode{"repo0": {
				ID:            "repoIdTEST",
				NameWithOwner: "org/some-repo-name",
			}},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockRepoJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "repoRequest with error",
			args: args{
				queryString: "", client: client,
			},
			want:               map[string]RepositoriesNode{"repo0": {}},
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
