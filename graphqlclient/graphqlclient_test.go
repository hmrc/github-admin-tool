package graphqlclient

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestClient_run_retries(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockHTTPResponder(
		"POST",
		"https://api.github.com/graphql",
		"",
		200,
	)

	r := &reportAccessService{}

	dryRun = tt.dryRunValue
	config.Team = tt.teamValue

	ctx := context.Background()

	request := NewRequest("some request")

	var respData interface{}

	client := NewClient()

	if err := client.Run(ctx, request, &respData); err != nil {
		t.Errorf("Run() error = %v", err)
	}

}
