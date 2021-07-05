# github-admin-tool - WORK IN PROGRESS

This is a tool to generate a repository report for an organization.

This tool aims to become a command line tool for adminstrative tasks in github for any given organization.  A bulk updater.

## Config

Please set the following ENV vars or use config.yml.example->config.yaml as file.

```bash
GHTOOL_TOKEN=github-bearer-token
ORG=github-org-name
```

## Help

As with any cli tool just run the following to see available actions/arguments.

`go run main.go -h`

## Report

Run the following command generate a CSV with respository settings.

`go run main.go report`

## Signing

Run the following command to turn commit signing on for all branch protection rules for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.

If the default branch does not have a protection rule, it will be created.

`go run main.go signing -r repos.txt`

## PR Approval

Run the following command to set pr-approval rules for a given branch name for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.  Check the command line help for different settings.

`go run main.go pr-approval -r repo_list.txt -b branch_name`
