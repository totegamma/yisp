package lib

import (
	"fmt"
	"strings"

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

	decoder := yaml.NewDecoder(strings.NewReader(yamlStr))
	if decoder == nil {
		return nil, core.NewEvaluationError(node, "failed to create yaml decoder")
	}

	var resultNodes []any
	for {
		var value any
		err := decoder.Decode(&value)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, core.NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to decode yaml: `%s`", yamlStr), err)
		}
		parsedNode, err := core.ParseAny(node.Attr.File(), value)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(node, "failed to parse yaml node", err)
		}
		resultNodes = append(resultNodes, parsedNode)
	}

	return &core.YispNode{
		Kind:  core.KindArray,
		Value: resultNodes,
		Attr:  node.Attr,
	}, nil
}
