package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"strings"
)

type reportAccess interface {
	getReport() (map[string]string, error)
}

type reportAccessService struct{}

func (r *reportAccessService) getReport() (map[string]string, error) {
	var (
		cursor           *string
		totalRecordCount int
		iteration        int
		bar              progressbar.Bar
	)

	client := graphqlclient.NewClient("https://api.github.com/graphql")
	query := reportAccessQuery()
	req := reportRequest(query)
	ctx := context.Background()
	iteration = 0

	adminAccess := make(map[string]string)

	if config.Team == "" || dryRun {
		return adminAccess, nil
	}

	for {
		// Set new cursor on every loop to paginate through 100 at a time
		req.Var("after", cursor)
		req.Var("team", config.Team)

		var respData OrganizationTeamsReponse
		if err := client.Run(ctx, req, &respData); err != nil {
			return adminAccess, fmt.Errorf("graphql call: %w", err)
		}

		teamNode := respData.OrganizationTeams.Teams.TeamNodes[0]

		cursor = &teamNode.TeamRepositories.PageInfo.EndCursor
		totalRecordCount = teamNode.TeamRepositories.TotalCount

		for _, value := range teamNode.TeamRepositories.Edges {
			adminAccess[value.Node.Name] = value.Permission
		}

		if iteration == 0 {
			bar.NewOption(0, totalRecordCount)
		}

		bar.Play(iteration)

		iteration += IterationCount

		if !teamNode.TeamRepositories.PageInfo.HasNextPage {
			bar.Play(totalRecordCount)

			break
		}
	}

	bar.Finish("Get team permissions")

	return adminAccess, nil
}

func reportAccessQuery() string {
	var query strings.Builder

	query.WriteString("query ($org: String! $after: String $team: String!) {")
	query.WriteString("		organization(login:$org) {")
	query.WriteString("			teams(query:$team, first: 1) {")
	query.WriteString("				nodes {")
	query.WriteString("					repositories(first: 100, after: $after) {")
	query.WriteString("						totalCount")
	query.WriteString("						pageInfo {")
	query.WriteString("							endCursor")
	query.WriteString("							hasNextPage")
	query.WriteString("						}")
	query.WriteString("						edges {")
	query.WriteString("							node {")
	query.WriteString("								name")
	query.WriteString("							}")
	query.WriteString("							permission")
	query.WriteString("						}")
	query.WriteString("					}")
	query.WriteString("				}")
	query.WriteString("			}")
	query.WriteString("		}")
	query.WriteString("	}")

	return query.String()
}
