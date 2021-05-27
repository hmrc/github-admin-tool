module github-admin-tool

go 1.16

replace github-admin-tool/cmd => ./cmd

replace github-admin-tool/graphqlclient => ./graphqlclient

require (
	github.com/jarcoal/httpmock v1.0.8
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
)
