package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/totegamma/yisp/pkg"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the yaml file",
	Long:  `Build the yaml file from the yisp script`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		allowCmd, err := cmd.Flags().GetBool("allow-cmd")
		if err == nil {
			yisp.SetAllowCmd(allowCmd)
		}

		yamlFile := args[0]
		if yamlFile == "" {
			cmd.Help()
			return
		}

		result, err := yisp.EvaluateYisp(yamlFile)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolP("allow-cmd", "", false, "Allow command execution")
}
