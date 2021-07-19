# github-admin-tool

This is a CLI tool used to audit/report on repositories, update branch protection signing and pr-approval settings for a given organisation.

By default it runs in a dry run mode.  Turn this off by adding `--dry-run=false` to any command.

## Installation

```bash
wget -O- https://github.com/hmrc/github-admin-tool/releases/download/v0.1.4/github-admin-tool_0.1.4_<OS_VERSION>.tar.gz | tar -xzv && chmod 755 github-admin-tool
```

## Config

Please set the following ENV vars or use config.yml.example->config.yaml as file.

github-bearer-token (PAT) should have the following: admin:org, repo, user.

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

`./github-admin-tool signing -r repo_list.txt`

## PR Approval

Run the following command to set pr-approval rules for a given branch name for the repos contained in the list.   The list should be a text file with repository names (without owner name) on new lines.  Check the command line help for different settings.

If the passed in branch does not have a protection rule, it will be created.

`./github-admin-tool pr-approval -r repo_list.txt -b branch_name`
