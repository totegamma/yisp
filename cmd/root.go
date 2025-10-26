package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "yisp",
	Short: "yisp is a command line tool for evaluating yisp expressions",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("cache-dir", "", "Directory to use for caching schemas and other data")
	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default is $HOME/.config/yisp/config.yaml)")
	cobra.OnInitialize(initConfig)
}

func initConfig() {

	configPath, _ := rootCmd.PersistentFlags().GetString("config")

	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		configDir := filepath.Join(home, ".config", "yisp")
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			err := os.MkdirAll(configDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}

		configPath = filepath.Join(configDir, "config.yaml")
	}

	viper.SetConfigFile(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		viper.SetDefault("AllowedGoPkgs", []string{})
		_ = viper.WriteConfig()
	}
}
