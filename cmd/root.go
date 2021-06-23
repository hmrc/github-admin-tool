package cmd

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const maxRepositories = 100

var (
	configFile             string // nolint // needed for cobra
	reposFile              string // nolint // needed for cobra
	config                 Config // nolint // using with viper
	dryRun                 bool   // nolint // using for global flag
	errInvalidRepo         = errors.New("invalid repo name")
	doBranchProtectionSend = branchProtectionSend // nolint // Like this for testing mock
	rootCmd                = &cobra.Command{      // nolint // needed for cobra
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
