package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("types", "assert", opAssertType)
	register("types", "get", opGetType)
	register("types", "of", opTypeOf)
}

func opAssertType(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("assert-type requires 2 arguments, got %d", len(cdr)))
	}

	schemaNode := cdr[0]
	valueNode := cdr[1]

	if schemaNode.Kind != core.KindType {
		return nil, core.NewEvaluationError(schemaNode, fmt.Sprintf("assert-type requires a type as the first argument, got %v", schemaNode.Kind))
	}
	schema, ok := schemaNode.Value.(*core.Schema)
	if !ok {
		return nil, core.NewEvaluationError(schemaNode, fmt.Sprintf("invalid type value: %T", schemaNode.Value))
	}

	err := schema.Validate(valueNode)
	if err != nil {
		return nil, core.NewEvaluationErrorWithParent(valueNode, fmt.Sprintf("value does not match schema: %s", err.Error()), err)
	}

	return valueNode, nil
}

func opGetType(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("get-type requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]

	if node.Type == nil {
		return &core.YispNode{
			Kind:  core.KindNull,
			Value: nil,
			Attr:  node.Attr,
		}, nil
	}

	return node.Type.ToYispNode()
}

func opTypeOf(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("typeof requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]

	if node.Type == nil {
		return &core.YispNode{
			Kind:  core.KindNull,
			Value: nil,
			Attr:  node.Attr,
		}, nil
	}

	return &core.YispNode{
		Kind:  core.KindType,
		Value: node.Type,
		Attr:  node.Attr,
	}, nil
}
