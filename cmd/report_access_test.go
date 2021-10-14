package cmd

import (
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func Test_reportAccessService_getReport(t *testing.T) {
	originalDryRun := dryRun
	originalTeamValue := config.Team

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	defer func() {
		dryRun = originalDryRun
		config.Team = originalTeamValue
	}()

	tests := []struct {
		name               string
		r                  *reportAccessService
		want               map[string]string
		wantErr            bool
		dryRunValue        bool
		teamValue          string
		mockHTTPReturnFile string
		mockHTTPStatus     int
	}{
		{
			name:        "getReport no config.team returns blank",
			want:        make(map[string]string),
			wantErr:     false,
			dryRunValue: false,
		},
		{
			name:        "getReport dry run on",
			want:        make(map[string]string),
			wantErr:     false,
			dryRunValue: true,
		},
		{
			name:               "getReport response error",
			want:               make(map[string]string),
			wantErr:            true,
			teamValue:          "some-org",
			dryRunValue:        false,
			mockHTTPReturnFile: "../testdata/mockAccessResponseError.json",
			mockHTTPStatus:     400,
		},
		{
			name: "getReport response success",
			want: map[string]string{
				"some-repo-1": "ADMIN",
				"some-repo-2": "WRITE",
				"some-repo-3": "ADMIN",
				"some-repo-4": "READ",
			},
			wantErr:            false,
			teamValue:          "some-org",
			dryRunValue:        false,
			mockHTTPReturnFile: "../testdata/mockAccessResponse.json",
			mockHTTPStatus:     200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockHTTPReturnFile != "" {
				mockHTTPResponder("POST", "https://api.github.com/graphql", tt.mockHTTPReturnFile, tt.mockHTTPStatus)
			}

			r := &reportAccessService{}

			dryRun = tt.dryRunValue
			config.Team = tt.teamValue
			got, err := r.getReport()
			if (err != nil) != tt.wantErr {
				t.Errorf("reportAccessService.getReport() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reportAccessService.getReport() = %v, want %v", got, tt.want)
			}
		})
	}
}
