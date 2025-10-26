package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/internal/yaml"
)

func init() {
	register("yaml", "marshal", opMarshal)
	register("yaml", "unmarshal", opUnmarshal)
}

func opMarshal(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("toYaml requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	yamlBytes, err := e.Render(node)
	if err != nil {
		return nil, core.NewEvaluationErrorWithParent(node, "failed to render yaml", err)
	}
	yamlStr := string(yamlBytes)

	return &core.YispNode{
		Kind:  core.KindString,
		Value: yamlStr,
		Attr:  node.Attr,
	}, nil
}

func opUnmarshal(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("fromYaml requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	if node.Kind != core.KindString {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("fromYaml requires a string argument, got %v", node.Kind))
	}

	yamlStr, ok := node.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid yaml string value: %T", node.Value))
	}

	var result any
	err := yaml.Unmarshal([]byte(yamlStr), &result)
	if err != nil {
		return nil, core.NewEvaluationErrorWithParent(node, "failed to unmarshal yaml", err)
	}

	resultNode, err := core.ParseAny(node.Attr.File(), result)
	if err != nil {
		return nil, core.NewEvaluationErrorWithParent(node, "failed to parse yaml result", err)
	}

	return resultNode, nil
}
