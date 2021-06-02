# github-admin-tool - WORK IN PROGRESS

This is a tool to generate a repository report for an organization.

This tool aims to become a command line tool for adminstrative tasks in github for a given organization.  A bulk updater.

# Config

Please set the following ENV vars or use config.yml.example->config.yaml as file.

```
GHTOOL_TOKEN=github-bearer-token
ORG=github-org-name
```

# Help

As with any cli tool just run the following to see available actions/arguments.

`go run main.go -h`

# Report

Run the following command generate a CSV with respository settings.

`go run main.go report`
