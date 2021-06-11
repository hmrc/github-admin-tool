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

var (
	configFile       string // nolint // needed for cobra
	reposFile        string // nolint // needed for cobra
	config           Config // nolint // using with viper
	dryRun           bool   // nolint // using for global flag
	errInvalidRepo   = errors.New("invalid repo name")
	signingCreate    = createSigningBranchProtection    // nolint // Like this for testing mock
	signingUpdate    = updateSigningBranchProtection    // nolint // Like this for testing mock
	prApprovalCreate = createPrApprovalBranchProtection // nolint // Like this for testing mock
	prApprovalUpdate = updatePrApprovalBranchProtection // nolint // Like this for testing mock
	rootCmd          = &cobra.Command{                  // nolint // needed for cobra
		Use:   "github-admin-tool",
		Short: "Github admin tool allows you to perform actions on your github repos",
		Long:  "Using Github version 4 GraphQL API to generate repo reports and administer your organisations repos etc",
	}
)

type Config struct {
	Token string `mapstructure:"token"`
	Org   string `mapstructure:"org"`
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
