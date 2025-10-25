package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"

	"github.com/totegamma/yisp/engine"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the yaml file",
	Long:  `Build the yaml file from the yisp script`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		showTrace, _ := cmd.Flags().GetBool("show-trace")
		renderSpecialObjects, _ := cmd.Flags().GetBool("render-special-objects")
		renderSourceMap, _ := cmd.Flags().GetBool("enable-sourcemap")
		allowUntypedManifest, _ := cmd.Flags().GetBool("allow-untyped-manifest")

		e := engine.NewEngine(engine.Options{
			ShowTrace:            showTrace,
			RenderSpecialObjects: renderSpecialObjects,
			RenderSources:        renderSourceMap,
			AllowUntypedManifest: allowUntypedManifest,
		})

		allowCmd, err := cmd.Flags().GetBool("allow-cmd")
		if err == nil {
			e.SetOption("net.gammalab.yisp.exec.allow_cmd", allowCmd)
		}

		allowedGoPkgs := viper.GetStringSlice("AllowedGoPkgs")
		if err == nil {
			e.SetOption("net.gammalab.yisp.exec.allowed_go_pkgs", allowedGoPkgs)
		}

		output, err := cmd.Flags().GetString("output")
		if err != nil {
			output = "yaml"
		}

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

		switch output {
		case "yaml":
			result, err := e.EvaluateFileToYaml(yamlFile)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Println(result)
		case "json":
			resultAny, err := e.EvaluateFileToAny(yamlFile)
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
		default:
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
	buildCmd.Flags().BoolP("enable-sourcemap", "", false, "Enable source map comments in output YAML")
	buildCmd.Flags().BoolP("allow-untyped-manifest", "", false, "Allow untyped manifest")
	buildCmd.Flags().StringP("output", "o", "yaml", "Output format (yaml, json)")
}
