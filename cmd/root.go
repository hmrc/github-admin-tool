package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  Config
	rootCmd = &cobra.Command{
		Use:   "github-admin-tool",
		Short: "Github admin tool allows you to perform actions on your github repos",
		Long:  "Using Github version 4 GraphQL API to generate repo reports and administer your organisations repos etc",
	}
)

type Config struct {
	Client struct {
		Token string `mapstructure:"token"`
		Org   string `mapstructure:"org"`
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default config.yaml)")
	rootCmd.PersistentFlags().Bool("dry-run", true, "dry-run mode to test command line options")
}

func initConfig() {
	// Try ENV vars
	viper.SetConfigType("env")
	viper.SetEnvPrefix("ghtool")
	viper.BindEnv("token")
	viper.BindEnv("org")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
}
