package ratelimit

import (
	"context"
	"fmt"
	"github-admin-tool/restclient"
)

type RateResponse struct {
	Resources struct {
		Rest struct {
			Limit     int   `json:"limit"`
			Used      int   `json:"used"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"core"`
		Graphql struct {
			Limit     int   `json:"limit"`
			Used      int   `json:"used"`
			Remaining int   `json:"remaining"`
			Reset     int64 `json:"reset"`
		} `json:"graphql"`
	} `json:"resources"`
}

func GetRateLimit(token string) (RateResponse, error) {
	client := restclient.NewClient("/rate_limit", token)
	response := RateResponse{}

	if err := client.Run(context.Background(), &response); err != nil {
		return response, fmt.Errorf("failed to retrieve rate limit data: %w", err)
	}

	return response, nil
}
