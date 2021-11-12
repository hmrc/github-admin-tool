# github-admin-tool

This is a CLI tool used to audit/report on repositories, update branch protection signing and pr-approval settings for a given organisation.

By default it runs in a dry run mode.  Turn this off by adding `--dry-run=false` to any command.

## Installation

1. Download the [latest](https://github.com/hmrc/github-admin-tool/releases/latest) archive for your OS. Older releases
[here](https://github.com/hmrc/github-admin-tool/releases)
2. Extract the binary e.g. `tar xvzf <archive>`
3. Make it executable e.g. `chmod +x github-admin-tool`

## Config

Set the following in config.yml ([template](config.yml.example)) OR set them as
environment variables, upper cased and prefixed with `GHTOOL_`

* token: (required) your GitHub personal access token  
           required scopes: admin:org, repo, user
* org:   (required) the GitHub organisation that you will scan
* team:  (optional) when specified will return the permissions that this team has on the repository

```bash
GHTOOL_TOKEN=token
GHTOOL_ORG=org
GHTOOL_TEAM=team
```

## Help

As with any cli tool just run the following to see available actions/arguments.

`./github-admin-tool -h`

## Repository Report

Run the following command generate a CSV report with respository settings and branch protection rules.

`./github-admin-tool report`

## Signing

Run the following command to turn commit signing on for all branch protection rules for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.

If the default branch does not have a protection rule, it will be created.

`./github-admin-tool signing -r repo_list.txt`

## PR Approval

Run the following command to set pr-approval rules for a given branch name for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.  Check the command line help for different settings.

If the passed in branch does not have a protection rule, it will be created.

`./github-admin-tool pr-approval -r repo_list.txt -b branch_name`
