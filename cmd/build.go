package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/totegamma/yisp/pkg"
	"path/filepath"
	"strings"
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

		showTrace, err := cmd.Flags().GetBool("show-trace")
		if err == nil {
			yisp.SetShowTrace(showTrace)
		}

		renderSpecialObjects, err := cmd.Flags().GetBool("render-special-objects")
		if err == nil {
			yisp.SetRenderSpecialObjects(renderSpecialObjects)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			output = "yaml"
		}

		allowedGoPkgs := viper.GetStringSlice("AllowedGoPkgs")
		yisp.SetAllowedPkgs(allowedGoPkgs)

		yamlFile := args[0]
		if yamlFile == "" {
			cmd.Help()
			return
		}

		if !strings.HasPrefix(yamlFile, "http://") && !strings.HasPrefix(yamlFile, "https://") {
			yamlFile, err = filepath.Abs(yamlFile)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}

		if output == "yaml" {
			result, err := yisp.EvaluateFileToYaml(yamlFile)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Println(result)
		} else if output == "json" {
			resultAny, err := yisp.EvaluateFileToAny(yamlFile)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			resultArr, ok := resultAny.([]any)
			if !ok {
				fmt.Println("Error: Result is not an array")
				return
			}

			if len(resultArr) == 0 {
				fmt.Println("Error: Result is empty")
				return
			}

			if len(resultArr) > 1 {
				fmt.Printf("Error: Json output only supports a single document, but got %v objects\n", len(resultArr))
				return
			}

			jsonResult, err := json.MarshalIndent(resultArr[0], "", "  ")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Println(string(jsonResult))
		} else {
			fmt.Println("Error: Unsupported output format. Use 'yaml' or 'json'.")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().BoolP("allow-cmd", "", false, "Allow command execution")
	buildCmd.Flags().BoolP("show-trace", "", false, "Show trace")
	buildCmd.Flags().BoolP("render-special-objects", "", false, "Show special objects (e.g. type, lambda, etc.)")
	buildCmd.Flags().StringP("output", "o", "yaml", "Output format (yaml, json)")
}
