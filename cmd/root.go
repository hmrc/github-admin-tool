package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github-admin-tool/graphqlclient"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const maxRepositories = 100

var (
	configFile                 string // nolint // needed for cobra
	reposFile                  string // nolint // needed for cobra
	config                     Config // nolint // using with viper
	dryRun                     bool   // nolint // using for global flag
	errInvalidRepo             = errors.New("invalid repo name")
	createBranchProtectionRule = createBranchProtection // nolint // Like this for testing mock
	updateBranchProtectionRule = updateBranchProtection // nolint // Like this for testing mock
	prApprovalCreate           = createBranchProtection // nolint // Like this for testing mock
	prApprovalUpdate           = updateBranchProtection // nolint // Like this for testing mock
	rootCmd                    = &cobra.Command{        // nolint // needed for cobra
		Use:   "github-admin-tool",
		Short: "Github admin tool allows you to perform actions on your github repos",
		Long:  "Using Github version 4 GraphQL API to generate repo reports and administer your organisations repos etc",
	}
)

type Config struct {
	Token string `mapstructure:"token"`
	Org   string `mapstructure:"org"`
}

type BranchProtectionArgs struct {
	Name     string
	DataType string
	Value    interface{}
}

func Execute() error {
	return errors.Wrap(rootCmd.Execute(), "root execute")
}

// nolint // needed for cobra
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default config.yaml)")
	rootCmd.PersistentFlags().Bool("dry-run", true, "dry-run mode to test command line options")
}

func initConfig() {
	var err error

	viper.SetConfigType("env")
	viper.SetEnvPrefix("ghtool")

	if err = viper.BindEnv("token"); err != nil {
		panic(fmt.Errorf("fatal error binding var: %w", err))
	}

	if err = viper.BindEnv("org"); err != nil {
		panic(fmt.Errorf("fatal error binding var: %w", err))
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigFile(configFile)

	if err = viper.ReadInConfig(); err != nil {
		log.Print("Could not find any config files")
	}

	if err = viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func readRepoList(reposFile string) ([]string, error) {
	var repos []string

	validRepoName := regexp.MustCompile("^[A-Za-z0-9_.-]+$")

	file, err := os.Open(reposFile)
	if err != nil {
		return repos, fmt.Errorf("could not open repo file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		repoName := scanner.Text()
		if !validRepoName.MatchString(repoName) {
			return repos, fmt.Errorf("%w: %s", errInvalidRepo, repoName)
		}

		repos = append(repos, repoName)
	}

	return repos, nil
}

func generateRepoQuery(repos []string) string {
	var signingQueryStr strings.Builder

	signingQueryStr.WriteString("fragment repoProperties on Repository {")
	signingQueryStr.WriteString("	id")
	signingQueryStr.WriteString("	nameWithOwner")
	signingQueryStr.WriteString("	description")
	signingQueryStr.WriteString("	defaultBranchRef {")
	signingQueryStr.WriteString("		name")
	signingQueryStr.WriteString("	}")
	signingQueryStr.WriteString("	branchProtectionRules(first: 100) {")
	signingQueryStr.WriteString("		nodes {")
	signingQueryStr.WriteString("			id")
	signingQueryStr.WriteString("			requiresCommitSignatures")
	signingQueryStr.WriteString("			pattern")
	signingQueryStr.WriteString("			requiresApprovingReviews")
	signingQueryStr.WriteString("			requiresCodeOwnerReviews")
	signingQueryStr.WriteString("			requiredApprovingReviewCount")
	signingQueryStr.WriteString("		}")
	signingQueryStr.WriteString("	}")
	signingQueryStr.WriteString("}")
	signingQueryStr.WriteString("query ($org: String!) {")

	for i := 0; i < len(repos); i++ {
		signingQueryStr.WriteString(fmt.Sprintf("repo%d: repository(owner: $org, name: \"%s\") {", i, repos[i]))
		signingQueryStr.WriteString("	...repoProperties")
		signingQueryStr.WriteString("}")
	}

	signingQueryStr.WriteString("}")

	return signingQueryStr.String()
}

func repoRequest(queryString string, client *graphqlclient.Client) (map[string]RepositoriesNodeList, error) {
	authStr := fmt.Sprintf("bearer %s", config.Token)

	req := graphqlclient.NewRequest(queryString)
	req.Var("org", config.Org)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", authStr)

	ctx := context.Background()

	var respData map[string]RepositoriesNodeList

	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, fmt.Errorf("graphql call: %w", err)
	}

	return respData, nil
}

func updateBranchProtection(
	branchProtectionRuleID string,
	branchProtectionArgs []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	var mutation, input, output strings.Builder

	mutation.WriteString("	mutation UpdateBranchProtectionRule(")
	mutation.WriteString("		$branchProtectionRuleId: String!,")
	mutation.WriteString("		$clientMutationId: String!,")

	input.WriteString("	updateBranchProtectionRule(")
	input.WriteString("		input:{")
	input.WriteString("			clientMutationId: $clientMutationId,")
	input.WriteString("			branchProtectionRuleId: $branchProtectionRuleId,")

	mutationBlock, inputBlock, requestVars := createQueryBlocks(branchProtectionArgs)

	mutation.WriteString(mutationBlock.String())
	mutation.WriteString("){")

	input.WriteString(inputBlock.String())
	input.WriteString("})")

	output.WriteString("{")
	output.WriteString("	branchProtectionRule {")
	output.WriteString("		id")
	output.WriteString("	}")
	output.WriteString("}}")

	req := graphqlclient.NewRequest(mutation.String() + input.String() + output.String())
	req.Var("clientMutationId", fmt.Sprintf("github-tool-%v", branchProtectionRuleID))
	req.Var("branchProtectionRuleId", branchProtectionRuleID)

	for key, value := range requestVars {
		req.Var(key, value)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	ctx := context.Background()

	if err := client.Run(ctx, req, nil); err != nil {
		return fmt.Errorf("from API call: %w", err)
	}

	return nil
}

func createQueryBlocks(branchProtectionArgs []BranchProtectionArgs) (
	mutation, input strings.Builder,
	requestVars map[string]interface{},
) {
	requestVars = make(map[string]interface{})

	for _, bprs := range branchProtectionArgs {
		mutation.WriteString(fmt.Sprintf("$%s: %s!,", bprs.Name, bprs.DataType))
		input.WriteString(fmt.Sprintf("%s: $%s,", bprs.Name, bprs.Name))
		requestVars[bprs.Name] = bprs.Value
	}

	return mutation, input, requestVars
}

func createBranchProtection(
	repositoryID,
	branchName string,
	branchProtectionArgs []BranchProtectionArgs,
	client *graphqlclient.Client,
) error {
	var mutation, input, output strings.Builder

	mutation.WriteString("mutation CreateBranchProtectionRule(")
	mutation.WriteString("	$repositoryId: String!,")
	mutation.WriteString("	$clientMutationId: String!,")
	mutation.WriteString("	$pattern: String!,")

	input.WriteString("		createBranchProtectionRule(")
	input.WriteString("			input:{")
	input.WriteString("				clientMutationId: $clientMutationId,")
	input.WriteString("				repositoryId: $repositoryId,")
	input.WriteString("				pattern: $pattern,")

	mutationBlock, inputBlock, requestVars := createQueryBlocks(branchProtectionArgs)

	mutation.WriteString(mutationBlock.String())
	mutation.WriteString(") {")

	input.WriteString(inputBlock.String())
	input.WriteString("})")

	output.WriteString("{")
	output.WriteString("	branchProtectionRule {")
	output.WriteString("		id")
	output.WriteString("	}")
	output.WriteString("}}")

	req := graphqlclient.NewRequest(mutation.String() + input.String() + output.String())
	req.Var("clientMutationId", fmt.Sprintf("github-tool-%v", repositoryID))
	req.Var("repositoryId", repositoryID)
	req.Var("pattern", branchName)

	for key, value := range requestVars {
		req.Var(key, value)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", config.Token))

	ctx := context.Background()

	if err := client.Run(ctx, req, nil); err != nil {
		return fmt.Errorf("from API call: %w", err)
	}

	return nil
}

func applyBranchProtection(
	repoSearchResult map[string]RepositoriesNodeList,
	action string,
	branchProtectionArgs []BranchProtectionArgs,
	client *graphqlclient.Client,
) (
	modified,
	created,
	info,
	problems []string,
) {
	var err error

OUTER:

	for _, repository := range repoSearchResult { // nolint
		if repository.DefaultBranchRef.Name == "" {
			info = append(info, fmt.Sprintf("No default branch for %v", repository.NameWithOwner))

			continue OUTER
		}

		// Check all nodes for default branch protection rule
		for _, branchProtection := range repository.BranchProtectionRules.Nodes {
			if repository.DefaultBranchRef.Name != branchProtection.Pattern {
				continue
			}

			// If default branch has already got signing turned on, no need to update
			if branchProtection.RequiresCommitSignatures {
				info = append(info, fmt.Sprintf("%s already turned on for %v", action, repository.NameWithOwner))

				continue OUTER
			}

			if err = updateBranchProtectionRule(branchProtection.ID, branchProtectionArgs, client); err != nil {
				problems = append(problems, err.Error())

				continue OUTER
			}
			modified = append(modified, repository.NameWithOwner)

			continue OUTER
		}

		if err = createBranchProtectionRule(
			repository.ID,
			repository.DefaultBranchRef.Name,
			branchProtectionArgs,
			client,
		); err != nil {
			problems = append(problems, err.Error())

			continue OUTER
		}

		created = append(created, repository.NameWithOwner)
	}

	return modified, created, info, problems
}
