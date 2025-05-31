package yisp

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/totegamma/yisp/internal/yaml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func getWholeYamlDocument(str string) (any, error) {

	decoder := yaml.NewDecoder(strings.NewReader(str))
	if decoder == nil {
		return nil, errors.New("failed to create decoder")
	}

	documents := make([]any, 0)
	for {
		var obj any
		err := decoder.Decode(&obj)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		documents = append(documents, obj)
	}

	return documents, nil
}

func TestYisp(t *testing.T) {

	files, err := filepath.Glob("../testdata/*.test.yaml")
	if err != nil {
		t.Fatalf("Error finding test files: %v", err)
	}

	for _, file := range files {

		expectedFile := strings.Replace(file, ".test.", ".expected.", 1)

		t.Run(file, func(t *testing.T) {

			abspath, err := filepath.Abs(file)
			if err != nil {
				t.Fatalf("Error getting absolute path for file %s: %v", file, err)
			}

			renderedStr, err := EvaluateYisp(abspath)
			if err != nil {
				t.Fatalf("Error evaluating Yisp file %s: %v", file, err)
			}

			rendered, err := getWholeYamlDocument(renderedStr)
			if err != nil {
				t.Fatalf("Error unmarshalling rendered string: %v", err)
			}

			expectedStr, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Fatalf("Error reading expected file %s: %v", expectedFile, err)
			}

			expected, err := getWholeYamlDocument(string(expectedStr))
			if err != nil {
				t.Fatalf("Error unmarshalling expected file: %v", err)
			}

			if !assert.Equal(t, expected, rendered) {
				t.Errorf("Rendered output does not match expected output for file %s", file)
			}
		})
	}

}
