package yisp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetK8sSchema(t *testing.T) {
	deploy, err := GetK8sSchema("apps", "v1", "Deployment")
	if assert.NoError(t, err) {
		assert.NotNil(t, deploy)
	}

	pod, err := GetK8sSchema("", "v1", "Pod")
	if assert.NoError(t, err) {
		assert.NotNil(t, pod)
	}
}
