package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("maps", "from-entries", opFromEntries)
	register("maps", "get", opGet)
	register("maps", "keys", opKeys)
	register("maps", "merge", opMerge)
	register("maps", "to-entries", opToEntries)
	register("maps", "values", opValues)
	register("maps", "make", opMakeMap)
}

func opFromEntries(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	node := cdr[0]
	arr, ok := node.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("fromEntries requires an array argument, got %v", node.Kind))
	}

	result := core.NewYispMap()
	for _, item := range arr {
		tupleNode, ok := item.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid tuple type: %T", item))
		}

		tupleArr, ok := tupleNode.Value.([]any)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid tuple value: %T", tupleNode.Value))
		}

		if len(tupleArr) != 2 {
			return nil, core.NewEvaluationError(node, "tuple must have exactly 2 elements")
		}

		keyNode, ok := tupleArr[0].(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid key type: %T", tupleArr[0]))
		}
		valueNode := tupleArr[1]

		keyStr, ok := keyNode.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(keyNode, fmt.Sprintf("invalid key value: %T", keyNode.Value))
		}

		result.Set(keyStr, valueNode)
	}

	return &core.YispNode{
		Kind:  core.KindMap,
		Value: result,
		Attr:  node.Attr,
	}, nil
}

func opGet(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("map requires 1 argument, got %d", len(cdr)))
	}

	mapValue, ok := cdr[0].Value.(*core.YispMap)
	if !ok {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("mapping-get requires a map argument, got %v", cdr[0].Kind))
	}

	keyValue, ok := cdr[1].Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(cdr[1], fmt.Sprintf("mapping-get requires a string key, got %v", cdr[1].Kind))
	}

	value, ok := mapValue.Get(keyValue)
	if !ok {
		return nil, core.NewEvaluationError(cdr[1], fmt.Sprintf("key %s not found in map", keyValue))
	}

	valueNode, ok := value.(*core.YispNode)
	if !ok {
		return nil, core.NewEvaluationError(cdr[1], fmt.Sprintf("invalid value type: %T", value))
	}

	return valueNode, nil
}

func opKeys(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("keys requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	mapValue, ok := node.Value.(*core.YispMap)
	if !ok {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("keys requires a map argument, got %v", cdr[0].Kind))
	}
	result := make([]any, 0)
	for key := range mapValue.AllFromFront() {
		result = append(result, &core.YispNode{
			Kind:  core.KindString,
			Value: key,
			Attr:  node.Attr,
		})
	}
	return &core.YispNode{
		Kind:  core.KindArray,
		Value: result,
		Attr:  node.Attr,
	}, nil
}

func opMerge(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	result := &core.YispNode{
		Kind: core.KindNull,
	}
	for _, node := range cdr {
		var err error
		result, err = core.DeepMergeYispNode(result, node, node.Type)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(node, "failed to merge map", err)
		}
	}

	return result, nil
}

func opToEntries(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("toEntries requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	mapValue, ok := node.Value.(*core.YispMap)
	if !ok {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("toEntries requires a map argument, got %v", cdr[0].Kind))
	}
	result := make([]any, 0)
	for key, value := range mapValue.AllFromFront() {
		tuple := &core.YispNode{
			Kind: core.KindArray,
			Value: []any{
				&core.YispNode{
					Kind:  core.KindString,
					Value: key,
					Attr:  node.Attr,
				},
				value,
			},
		}
		result = append(result, tuple)
	}
	return &core.YispNode{
		Kind:  core.KindArray,
		Value: result,
		Attr:  node.Attr,
	}, nil
}

func opValues(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("values requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	mapValue, ok := node.Value.(*core.YispMap)
	if !ok {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("values requires a map argument, got %v", cdr[0].Kind))
	}
	result := make([]any, 0)
	for _, value := range mapValue.AllFromFront() {
		result = append(result, value)
	}
	return &core.YispNode{
		Kind:  core.KindArray,
		Value: result,
		Attr:  node.Attr,
	}, nil
}

func opMakeMap(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	mapNode := core.NewYispMap()

	for i := 0; i < len(cdr); i += 2 {
		keyNode := cdr[i]
		if keyNode.Kind != core.KindString {
			return nil, core.NewEvaluationError(keyNode, fmt.Sprintf("map keys must be strings, got %v", keyNode.Kind))
		}
		keyStr, _ := keyNode.Value.(string)

		if i+1 >= len(cdr) {
			return nil, core.NewEvaluationError(keyNode, "map requires an even number of arguments")
		}
		valueNode := cdr[i+1]

		mapNode.Set(keyStr, valueNode)
	}

	return &core.YispNode{
		Kind:  core.KindMap,
		Value: mapNode,
	}, nil
}
