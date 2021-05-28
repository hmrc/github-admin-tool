package cmd

import (
	"github-admin-tool/graphqlclient"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

var client = graphqlclient.NewClient("https://api.github.com/graphql")

func Test_reportRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var (
		oneResult   []ReportResponse
		oneResponse ReportResponse
	)

	oneResponse.Organization.Repositories.TotalCount = 1
	nodes := make([]RepositoriesNodeList, 1)
	nodes[0].Name = "repo-name"
	nodes[0].NameWithOwner = "org-name/repo-name"
	nodes[0].DefaultBranchRef.Name = "master"
	nodes[0].SquashMergeAllowed = true
	oneResponse.Organization.Repositories.Nodes = nodes
	oneResult = append(oneResult, oneResponse)

	tests := []struct {
		name               string
		mockHTTPReturnFile string
		want               []ReportResponse
	}{
		{name: "reportRequestReturnsEmpty", mockHTTPReturnFile: "../testdata/mockEmptyJsonResponse.json", want: nil},
		{name: "reportRequestReturnsOne", mockHTTPReturnFile: "../testdata/mockJsonResponse.json", want: oneResult},
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
				httpmock.NewStringResponder(200, string(mockHTTPReturn)),
			)

			if got, err := reportRequest(client); !reflect.DeepEqual(got, tt.want) {
				if err != nil {
					t.Fatalf("failed to run reportRequest %v", err)
				}
				t.Errorf("reportRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
