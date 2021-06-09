package cmd

import (
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func Test_repoRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := graphqlclient.NewClient("https://api.github.com/graphql")

	wantOneNodeList := make(map[string]RepositoriesNodeList)
	wantOneNodeList["repo0"] = RepositoriesNodeList{
		ID:            "repoIdTEST",
		NameWithOwner: "org/some-repo-name",
	}

	emptyNodeList := make(map[string]RepositoriesNodeList)
	emptyNodeList["repo0"] = RepositoriesNodeList{}

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
			want:               wantOneNodeList,
			wantErr:            false,
			mockHTTPReturnFile: "../testdata/mockSigningJsonResponse.json",
			mockHTTPStatusCode: 200,
		},
		{
			name: "GetRepoResponseWithError",
			args: args{
				queryString: "", client: client,
			},
			want:               emptyNodeList,
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
