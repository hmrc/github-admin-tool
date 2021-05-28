package cmd

import (
	"reflect"
	"testing"

	"github-admin-tool/graphqlclient"

	"github.com/jarcoal/httpmock"
)

var client = graphqlclient.NewClient("https://api.github.com/graphql")

var mockJsonResponse string = `
{
	"data": {
	  "organization": {
		"repositories": {
		  "totalCount": 1,
		  "pageInfo": {
			"endCursor": "",
			"hasNextPage": false
		  },
		  "nodes": [
			{
			  "deleteBranchOnMerge": false,
			  "isArchived": false,
			  "isEmpty": false,
			  "isFork": false,
			  "isPrivate": false,
			  "mergeCommitAllowed": false,
			  "name": "repo-name",
			  "nameWithOwner": "org-name/repo-name",
			  "rebaseMergeAllowed": false,
			  "squashMergeAllowed": true,
			  "defaultBranchRef": {
				"name": "master"
			  },
			  "parent": null
			}
		  ]
		}
	  }
	}
  }
`

var mockEmptyJsonResponse string = `
{
	"data": {}
}
`

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
		name           string
		mockHttpReturn string
		want           []ReportResponse
	}{
		{name: "reportRequestReturnsEmpty", mockHttpReturn: mockEmptyJsonResponse, want: emptyAllResults},
		{name: "reportRequestReturnsOne", mockHttpReturn: mockJsonResponse, want: oneResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpmock.RegisterResponder(
				"POST",
				"https://api.github.com/graphql",
				httpmock.NewStringResponder(200, tt.mockHttpReturn),
			)

			if got, _ := reportRequest(client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
