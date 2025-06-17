package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

type open struct {
	Definitions map[string]map[string]any `json:"definitions"`
}

var cacheKubeSchemas = &cobra.Command{
	Use:   "cache-kube-schemas",
	Short: "Cache Kubernetes Schemas",
	Long:  `This command fetches the Kubernetes OpenAPI schema and saves it locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		response, err := cacheKube()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching Kubernetes : %v\n", err)
			os.Exit(1)
		}

		var openapi open
		err = json.Unmarshal([]byte(response), &openapi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error unmarshalling response: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully fetched Kubernetes . Saving schemas...")
		err = saveSchemas(openapi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving schemas: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Schemas saved successfully.")
	},
}

func cacheKube() (string, error) {

	cmd := exec.Command("kubectl", "get", "--raw", "/openapi/v2")

	fmt.Fprintf(os.Stderr, "Going to run command: %v\n", cmd.Args)
	fmt.Fprintf(os.Stderr, "Press Enter to continue or Ctrl+C to cancel...\n")
	_, err := os.Stdin.Read(make([]byte, 1))
	if err != nil {
		return "", fmt.Errorf("error reading input: %v", err)
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	errorOutput := stderr.String()
	if err != nil {
		return "", fmt.Errorf("error running command: %v, stderr: %s", err, errorOutput)
	}

	response := stdout.String()
	if response == "" {
		return "", fmt.Errorf("no response received from the Kubernetes ")
	}
	return response, nil
}

func saveSchemas(openapi open) error {

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	schemasPath := filepath.Join(home, ".cache", "yisp", "schemas")
	if err := os.MkdirAll(schemasPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating schemas directory: %v\n", err)
		os.Exit(1)
	}

	gvkPath := filepath.Join(home, ".cache", "yisp", "gvk")
	if err := os.MkdirAll(gvkPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating GVK directory: %v\n", err)
		os.Exit(1)
	}

	for key, def := range openapi.Definitions {

		def["$id"] = key

		schemaPath := filepath.Join(schemasPath, fmt.Sprintf("%s.json", key))

		file, err := os.Create(schemaPath)
		if err != nil {
			fmt.Println("Error creating file:", err)
			continue
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(def)
		if err != nil {
			fmt.Println("Error encoding JSON to file:", err)
			continue
		}
		fmt.Printf("Created file: %s\n", schemaPath)

		// Extract GVK from the key
		gvkAny, ok := def["x-kubernetes-group-version-kind"]
		if !ok {
			fmt.Println("No GVK found for key:", key)
			continue
		}
		gvks, ok := gvkAny.([]any)
		if !ok {
			fmt.Println("Invalid GVK format for key:", key)
			continue
		}

		for _, gvkAny := range gvks {
			gvkMap, ok := gvkAny.(map[string]any)
			if !ok {
				fmt.Println("Invalid GVK structure for key:", key)
				continue
			}

			group, ok := gvkMap["group"]
			if !ok {
				fmt.Println("No group found in GVK for key:", key)
				continue
			}
			version, ok := gvkMap["version"]
			if !ok {
				fmt.Println("No version found in GVK for key:", key)
				continue
			}
			kind, ok := gvkMap["kind"]
			if !ok {
				fmt.Println("No kind found in GVK for key:", key)
				continue
			}

			gvkPath := filepath.Join(gvkPath, fmt.Sprintf("%s_%s_%s.txt", group, version, kind))

			file, err = os.Create(gvkPath)
			if err != nil {
				fmt.Println("Error creating file:", err)
				continue
			}

			//write key
			_, err = file.WriteString(key)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				continue
			}
			fmt.Printf("Created GVK file: %s\n", gvkPath)

		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cacheKubeSchemas)
}
