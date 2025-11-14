package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/internal/yaml"
)

func init() {
	register("utils", "op-patch", opOpPatch)
}

type JsonPatch struct {
	Op    string `json:"op"`
	From  string `json:"from,omitempty"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

func opOpPatch(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("patch requires 2 arguments, got %d", len(cdr)))
	}

	// target := cdr[0]
	patchesNode := cdr[1]
	patchesYaml, err := e.Render(patchesNode)
	if err != nil {
		return nil, core.NewEvaluationError(patchesNode, "failed to render patches to yaml")
	}

	var patches []JsonPatch
	err = yaml.Unmarshal([]byte(patchesYaml), &patches)
	if err != nil {
		return nil, core.NewEvaluationError(patchesNode, "failed to parse patches yaml to JsonPatch objects: "+err.Error())
	}

	return nil, nil
}
