package graphqlclient

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestClient_run_retries(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockHTTPReturnFile := "testdata/mockTimeoutResponse.json"
	mockHTTPReturn, err := ioutil.ReadFile(mockHTTPReturnFile)

	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	httpmock.RegisterResponder(
		"POST",
		"https://api.github.com/graphql",
		httpmock.NewStringResponder(200, string(mockHTTPReturn)),
	)

	ctx := context.Background()

	request := NewRequest("some request")

	var respData interface{}

	client := NewClient()

	if err := client.Run(ctx, request, &respData); err != nil {
		t.Errorf("Run() error = %v", err)
	}

}
