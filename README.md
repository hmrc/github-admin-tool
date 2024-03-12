# github-admin-tool

This is a CLI tool used to :
* Audit/report on repositories
* Update branch protection signing and pr-approval settings for a given organisation
* Report on repository webhooks
* Removal of webhooks by hostname for a given list of reposiotories

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

Run the following command to generate a CSV or JSON report with respository settings and branch protection rules.

`./github-admin-tool report`

## Repository webhook report

Run the following command to generate a CSV or JSON report with respository webhook settings.

`./github-admin-tool report-webhook`

Note: You can use jq to generate a list of repositories containing a certain webhook with this command:

`jq 'to_entries | map(select(.value[].config.url | contains("WEBHOOK_URL"))) | map(.key)' github_webhook_report.json`

## Signing

Run the following command to turn commit signing on for all branch protection rules for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.

If the default branch does not have a protection rule, it will be created.

`./github-admin-tool signing -r repo_list.txt`

## PR approval

Run the following command to set pr-approval rules for a given branch name for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.  Check the command line help for different settings.

If the passed in branch does not have a protection rule, it will be created.

`./github-admin-tool pr-approval -r repo_list.txt -b branch_name`

## Webhook removal

Run the following command to remove a webhook for the repos contained in the given list and URL (full URL with protocol).   The list should be a text file with repository names (without owner name) on new lines.  Check the command line help for different settings.

`./github-admin-tool webhook-remove -r repo_list.txt -u webhook_url`

## Dependabot settings

Run the following command to modify the dependabot settings for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.  Check the command line help for different settings.

`./github-admin-tool dependabot -r repo_list.txt`

## Developer release

Use the go releaser tool to make a release to the repo.

<https://goreleaser.com/quick-start/>

```bash
export GITHUB_TOKEN=<ghp_xxxtokenxxx>
git tag -a v0.1.0 -m "First release"
git push origin v0.1.0

goreleaser release --clean
```
