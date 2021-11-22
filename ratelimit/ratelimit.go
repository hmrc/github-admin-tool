package ratelimit

import (
	"context"
	"fmt"
	"github-admin-tool/restclient"
	"net/http"
)

type RateResponse struct {
	Resources RateResources `json:"resources"`
}

type RateResources struct {
	Rest    RateRest    `json:"core"`
	Graphql RateGraphql `json:"graphql"`
}

type RateRest struct {
	Limit     int   `json:"limit"`
	Used      int   `json:"used"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
}

type RateGraphql struct {
	Limit     int   `json:"limit"`
	Used      int   `json:"used"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
}

func GetRateLimit(token string) (RateResponse, error) {
	client := restclient.NewClient("/rate_limit", token, http.MethodGet)
	response := RateResponse{}

	if err := client.Run(context.Background(), &response); err != nil {
		return response, fmt.Errorf("failed to retrieve rate limit data: %w", err)
	}

	return response, nil
}
