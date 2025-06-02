package main

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

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	configPath := filepath.Join(home, ".config", "yisp")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		err := os.MkdirAll(configPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	viper.SetConfigFile(filepath.Join(configPath, "config.yaml"))

	err = viper.ReadInConfig()
	if err != nil {
		viper.SetDefault("AllowedGoPkgs", []string{})
		_ = viper.WriteConfig()
	}
}
