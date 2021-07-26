package cmd

import (
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"github-admin-tool/progressbar"
	"log"
	"strings"
)

func reportAccessQuery() string {
	var query strings.Builder

	query.WriteString("query {")
	query.WriteString("		organization(login:$org) {")
	query.WriteString("			teams(query: \"Repository Admins\", first: 1) {")
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

func reportAccessGet() (map[string]string, error) {
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

	for {
		// Set new cursor on every loop to paginate through 100 at a time
		req.Var("after", cursor)

		var respData OrganizationTeamsReponse
		if err := client.Run(ctx, req, &respData); err != nil {
			return adminAccess, fmt.Errorf("graphql call: %w", err)
		}

		cursor = &respData.OrganizationTeams.Teams.TeamRepositories.PageInfo.EndCursor
		totalRecordCount = respData.OrganizationTeams.Teams.TeamRepositories.TotalCount

		if dryRun {
			log.Printf("This is a dry run, the report would process %d records\n", totalRecordCount)

			break
		}

		for _, value := range respData.OrganizationTeams.Teams.TeamRepositories.Edges {
			adminAccess[value.Node.Name] = value.Permission
		}

		if iteration == 0 {
			bar.NewOption(0, totalRecordCount)
		}

		if !respData.OrganizationTeams.Teams.TeamRepositories.PageInfo.HasNextPage {
			iteration = totalRecordCount
			bar.Play(iteration)

			break
		}

		bar.Play(iteration)

		iteration += IterationCount
	}

	bar.Finish()

	return adminAccess, nil
}
