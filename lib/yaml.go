package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("yaml", "marshal", opToYaml)
}

func opToYaml(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
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
