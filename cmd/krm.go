package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
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
}

var krmCmd = &cobra.Command{
	Use:  "krm",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		var krmInput KRMInput
		err = yaml.Unmarshal(input, &krmInput)
		if err != nil {
			panic(err)
		}

		e := engine.NewEngine(engine.Options{
			ShowTrace:            false,
			RenderSpecialObjects: false,
			RenderSources:        false,
			AllowUntypedManifest: krmInput.FunctionConfig.Spec.AllowUntypedManifest,
		})

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

			result, err := e.EvaluateFileToYaml(target)
			if err != nil {
				panic(err)
			}

			fmt.Println(result)

		} else {
			scriptReader := strings.NewReader(krmInput.FunctionConfig.Spec.YispScript)

			wd, err := os.Getwd()
			if err != nil {
				panic(err)
			}

			tmp := filepath.Join(wd, "stdin.krm.yaml")

			result, err := e.EvaluateReaderToYaml(scriptReader, tmp)
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
