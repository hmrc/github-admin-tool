package cmd

import (
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func Test_repositoryReaderService_read(t *testing.T) {
	type args struct {
		reposFile string
	}

	tests := []struct {
		name    string
		r       *repositoryReaderService
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "repositoryList returns error",
			args: args{
				reposFile: "blah.txt",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "repositoryList returns one repo",
			args: args{
				reposFile: "../testdata/one_repo_list.txt",
			},
			want:    []string{"a-test-repo"},
			wantErr: false,
		},
		{
			name: "repositoryList returns two repos",
			args: args{
				reposFile: "../testdata/two_repo_list.txt",
			},
			want:    []string{"a-test-repo", "a-test-repo2"},
			wantErr: false,
		},
		{
			name: "repositoryList throws error",
			args: args{
				reposFile: "../testdata/repo_list_bad.txt",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repositoryReaderService{}
			got, err := r.read(tt.args.reposFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("repositoryReaderService.read() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("repositoryReaderService.read() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_repositoryQuery(t *testing.T) {
	type args struct {
		repos []string
	}

	tests := []struct {
		name     string
		args     args
		filePath string
	}{
		{
			name: "repositoryQuery with one repo",
			args: args{
				repos: []string{"repo-name-1"},
			},
			filePath: "../testdata/mockGenerateRepoWithOneRepo.txt",
		},
		{
			name: "repositoryQuery with two repos",
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

			got := repositoryQuery(tt.args.repos)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("repositoryQuery() = %s, mockReturn %s'", got, want)
			}
		})
	}
}

func Test_repositoryRequest(t *testing.T) {
	type args struct {
		queryString string
	}

	tests := []struct {
		name      string
		args      args
		wantQuery string
		wantVars  map[string]interface{}
	}{
		{
			name: "repositoryRequest",
			args: args{
				queryString: "some query",
			},
			wantQuery: "some query",
			wantVars: map[string]interface{}{
				"org": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repositoryRequest(tt.args.queryString)

			if !reflect.DeepEqual(got.Query(), tt.wantQuery) {
				t.Errorf("repositoryRequest() query = %v, want %v", got.Query(), tt.wantQuery)
			}

			if !reflect.DeepEqual(got.Vars(), tt.wantVars) {
				t.Errorf("repositoryRequest() vars = %T, want %T", got.Vars(), tt.wantVars)
			}
		})
	}
}

func Test_repositoryGetterService_get(t *testing.T) {
	type args struct {
		repositoryList []string
		sender         *githubRepositorySender
	}

	tests := []struct {
		name             string
		r                *repositoryGetterService
		args             args
		wantRepositories map[string]*RepositoriesNode
		wantErr          bool
	}{
		{
			name: "repositoryGet returns error",
			args: args{
				repositoryList: []string{
					"repo-name1",
					"repo-name2",
				},
				sender: &githubRepositorySender{
					sender: &mockRepositorySender{sendFail: true},
				},
			},
			wantErr: true,
		},
		{
			name: "repositoryGet success",
			args: args{
				repositoryList: []string{
					"repo-name1",
					"repo-name2",
				},
				sender: &githubRepositorySender{
					sender: &mockRepositorySender{sendFail: false},
				},
			},
			wantRepositories: make(map[string]*RepositoriesNode),
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &repositoryGetterService{}
			gotRepositories, err := r.get(tt.args.repositoryList, tt.args.sender)
			if (err != nil) != tt.wantErr {
				t.Errorf("repositoryGetterService.get() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(gotRepositories, tt.wantRepositories) {
				t.Errorf("repositoryGetterService.get() = %v, want %v", gotRepositories, tt.wantRepositories)
			}
		})
	}
}

func Test_repositorySenderService_send(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		req *graphqlclient.Request
	}

	tests := []struct {
		name               string
		r                  *repositorySenderService
		args               args
		want               map[string]*RepositoriesNode
		wantErr            bool
		mockHTTPReturnFile string
		mockHTTPStatusCode int
	}{
		{
			name: "repositorySend success",
			args: args{
				req: graphqlclient.NewRequest("query"),
			},
			want: map[string]*RepositoriesNode{"repo0": {
				ID:            "repoIdTEST",
				NameWithOwner: "org/some-repo-name",
			}},
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockRepoJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "repositorySend failure",
			args: args{
				req: graphqlclient.NewRequest("query"),
			},
			wantErr:            true,
			mockHTTPReturnFile: "../testdata/mockEmptyResponse.json",
			mockHTTPStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPResponder("POST", "https://api.github.com/graphql", tt.mockHTTPReturnFile, tt.mockHTTPStatusCode)

			r := &repositorySenderService{}
			got, err := r.send(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("repositorySenderService.send() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("repositorySenderService.send() = %v, want %v", got, tt.want)
			}
		})
	}
}
