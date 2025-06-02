package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime/debug"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of the application",
	Long:  `All software has versions. This is the version of this application.`,
	Run: func(cmd *cobra.Command, args []string) {

		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("Build information is not available.")
			return
		}

		fmt.Printf("Version: %s\n", buildInfo.Main.Version)
		fmt.Printf("Go Version: %s\n", buildInfo.GoVersion)
		fmt.Printf("Build Settings:\n")
		for _, setting := range buildInfo.Settings {
			fmt.Printf("  %s: %s\n", setting.Key, setting.Value)
		}

	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
