# github-admin-tool - WORK IN PROGRESS

This is a tool to generate a repository report for an organization.

This tool aims to become a command line tool for adminstrative tasks in github for any given organization.  A bulk updater.

## Config

Please set the following ENV vars or use config.yml.example->config.yaml as file.

```bash
GHTOOL_TOKEN=github-bearer-token
GHTOOL_ORG=github-org-name
```

## Help

As with any cli tool just run the following to see available actions/arguments.

`./github-admin-tool -h`

## Report

Run the following command generate a CSV with respository settings.

`./github-admin-tool report`

## Signing

Run the following command to turn commit signing on for all branch protection rules for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.

If the default branch does not have a protection rule, it will be created.

`./github-admin-tool signing -r repos.txt`

## PR Approval

Run the following command to set pr-approval rules for a given branch name for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.  Check the command line help for different settings.

`./github-admin-tool pr-approval -r repo_list.txt -b branch_name`
