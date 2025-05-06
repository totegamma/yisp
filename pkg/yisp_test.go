package yisp

import (
	"github.com/stretchr/testify/assert"
	"github.com/totegamma/yisp/yaml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYisp(t *testing.T) {

	files, err := filepath.Glob("../testdata/test*")
	if err != nil {
		t.Fatalf("Error finding test files: %v", err)
	}

	for _, file := range files {

		expectedFile := strings.Replace(file, "test_", "expected_", 1)

		t.Run(file, func(t *testing.T) {

			renderedStr, err := EvaluateYisp(file)
			if err != nil {
				t.Fatalf("Error evaluating Yisp file %s: %v", file, err)
			}

			var rendered any
			err = yaml.Unmarshal([]byte(renderedStr), &rendered)
			if err != nil {
				t.Fatalf("Error unmarshalling rendered string: %v", err)
			}

			expectedStr, err := os.ReadFile(expectedFile)
			if err != nil {
				t.Fatalf("Error reading expected file %s: %v", expectedFile, err)
			}

			var expected any
			err = yaml.Unmarshal(expectedStr, &expected)
			if err != nil {
				t.Fatalf("Error unmarshalling expected file: %v", err)
			}

			if !assert.Equal(t, expected, rendered) {
				t.Errorf("Rendered output does not match expected output for file %s", file)
			}
		})
	}

}
