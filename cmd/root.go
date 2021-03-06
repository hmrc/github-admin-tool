package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile        string // nolint // needed for cobra
	reposFile         string // nolint // needed for cobra
	branchName        string // nolint // needed for cobra
	webhookURL        string // nolint // needed for cobra
	config            Config // nolint // using with viper
	dryRun            bool   // nolint // using for global flag
	ignoreArchived    bool   // nolint // modifying within this package
	filePath          string // nolint // modifying within this package
	fileType          string // nolint // modifying within this package
	errInvalidRepo    = errors.New("invalid repo name")
	errInvalidTimeout = errors.New("invalid timeout")
	rootCmd           = &cobra.Command{ // nolint // needed for cobra
		Use:   "github-admin-tool",
		Short: "Github admin tool allows you to perform actions on your github repos",
		Long: `Using Github GraphQL API where possible (some actions only available using REST API) 
		to generate repo reports and administer your organisations repos etc`,
	}
)

const (
	// IterationCount the number of repos per result set.
	IterationCount int = 100
)

type Config struct {
	Token string `mapstructure:"token"`
	Org   string `mapstructure:"org"`
	Team  string `mapstructure:"team"`
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
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

	if err = viper.BindEnv("team"); err != nil {
		panic(fmt.Errorf("fatal error binding var: %w", err))
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigFile(configFile)

	viper.ReadInConfig() // nolint // don't want to do anything here if no config

	if err = viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}
