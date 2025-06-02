package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var allowCmd = &cobra.Command{
	Use:   "allow",
	Short: "Add go pkg to allow list",
	Long:  `Add go pkg to allow list`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pkg := args[0]
		viper.Set(
			"AllowedGoPkgs",
			append(
				viper.GetStringSlice("AllowedGoPkgs"),
				pkg,
			),
		)
		err := viper.WriteConfig()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Added %s to allowed go pkgs\n", pkg)
	},
}

func init() {
	rootCmd.AddCommand(allowCmd)
}
