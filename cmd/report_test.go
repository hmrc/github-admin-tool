package cmd

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github-admin-tool/graphqlclient"

	"github.com/jarcoal/httpmock"
)

var client = graphqlclient.NewClient("https://api.github.com/graphql")

func Test_reportRequest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	var emptyAllResults []ReportResponse
	var oneResponse ReportResponse

	oneResponse.Organization.Repositories.TotalCount = 1
	nodes := make([]RepositoriesNodeList, 1)
	nodes[0].Name = "repo-name"
	nodes[0].NameWithOwner = "org-name/repo-name"
	nodes[0].DefaultBranchRef.Name = "master"
	nodes[0].SquashMergeAllowed = true
	oneResponse.Organization.Repositories.Nodes = nodes
	oneResult := append(emptyAllResults, oneResponse)

	tests := []struct {
		name               string
		mockHttpReturnFile string
		want               []ReportResponse
	}{
		{name: "reportRequestReturnsEmpty", mockHttpReturnFile: "../testdata/mockEmptyJsonResponse.json", want: emptyAllResults},
		{name: "reportRequestReturnsOne", mockHttpReturnFile: "../testdata/mockJsonResponse.json", want: oneResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpReturn, err := ioutil.ReadFile(tt.mockHttpReturnFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}
			httpmock.RegisterResponder(
				"POST",
				"https://api.github.com/graphql",
				httpmock.NewStringResponder(200, string(mockHttpReturn)),
			)

			if got, _ := reportRequest(client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
