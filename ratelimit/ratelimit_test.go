package ratelimit

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestGetRateLimit(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		token string
	}

	tests := []struct {
		name               string
		args               args
		mockHTTPReturnFile string
		want               RateResponse
		wantErr            bool
	}{
		{

			name: "GetRateLimit fail",
			args: args{
				token: "TOKEN",
			},
			mockHTTPReturnFile: "testdata/mockEmptyResponse.json",
			wantErr:            true,
		},
		{
			name: "GetRateLimit success",
			args: args{
				token: "TOKEN",
			},
			mockHTTPReturnFile: "testdata/mockRestRateLimitResponse.json",
			want: RateResponse{
				Resources: RateResources{
					Rest: RateRest{
						Limit:     5000,
						Used:      3,
						Remaining: 4997,
						Reset:     1633943925,
					},
					Graphql: RateGraphql{
						Limit:     5000,
						Used:      0,
						Remaining: 5000,
						Reset:     1633947046,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPReturn, err := ioutil.ReadFile(tt.mockHTTPReturnFile)
			if err != nil {
				t.Fatalf("failed to read test data: %v", err)
			}

			httpmock.RegisterResponder(
				"GET",
				"https://api.github.com/rate_limit",
				httpmock.NewStringResponder(200, string(mockHTTPReturn)),
			)

			got, err := GetRateLimit(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRateLimit() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRateLimit() = %v, want %v", got, tt.want)
			}
		})
	}
}
