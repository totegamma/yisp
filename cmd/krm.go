package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/engine"
	"github.com/totegamma/yisp/internal/yaml"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Metadata struct {
	Name        string            `yaml:"name"`
	Annotations map[string]string `yaml:"annotations"`
}

type Spec struct {
	AllowUntypedManifest bool   `yaml:"allowUntypedManifest"`
	YispScript           string `yaml:"yisp"`
	Target               string `yaml:"target"`
}

type FunctionConfig struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type KRMInput struct {
	ApiVersion     string         `yaml:"apiVersion"`
	Kind           string         `yaml:"kind"`
	FunctionConfig FunctionConfig `yaml:"functionConfig"`
	Items          []any          `yaml:"items"`
}

var krmCmd = &cobra.Command{
	Use:  "krm",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		inputStr, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		var krmInput KRMInput
		err = yaml.Unmarshal(inputStr, &krmInput)
		if err != nil {
			panic(err)
		}

		e := engine.NewEngine(engine.Options{
			ShowTrace:            false,
			RenderSpecialObjects: false,
			RenderSources:        false,
			AllowUntypedManifest: krmInput.FunctionConfig.Spec.AllowUntypedManifest,
		})

		env := core.NewEnv()

		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		var items []any
		for _, item := range krmInput.Items {
			apiVersion, ok := item.(map[string]any)["apiVersion"]
			if ok && apiVersion == "krm.yisp.gammalab.net/v1" {
				continue
			}
			items = append(items, item)
		}

		itemNodes, err := core.ParseAny(filepath.Join(wd, "items.krm.yaml"), items)
		if err != nil {
			panic(err)
		}
		itemNodes.IsDocumentRoot = true

		env.Set("items", itemNodes)

		if krmInput.FunctionConfig.Spec.YispScript == "" {

			target := krmInput.FunctionConfig.Spec.Target
			if target == "" {
				target = "."
			}

			if target != "-" && !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
				target, err = filepath.Abs(target)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
			}

			result, err := e.EvaluateFileToYamlWithEnv(target, env)
			if err != nil {
				panic(err)
			}

			fmt.Println(result)

		} else {

			scriptReader := strings.NewReader(krmInput.FunctionConfig.Spec.YispScript)

			tmp := filepath.Join(wd, "stdin.krm.yaml")

			result, err := e.EvaluateReaderToYamlWithEnv(scriptReader, env, tmp)
			if err != nil {
				panic(err)
			}

			fmt.Println(result)
		}

	},
}

func init() {
	rootCmd.AddCommand(krmCmd)
}
