package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("lists", "car", opCar)
	register("lists", "cdr", opCdr)
	register("lists", "cons", opCons)
	register("lists", "filter", opFilter)
	register("lists", "flatten", opFlatten)
	register("lists", "map", opMap)
	register("lists", "reduce", opReduce)
	register("lists", "iota", opIota)
}

// opCar returns the first element of a list
func opCar(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("car requires 1 argument, got %d", len(cdr)))
	}

	listNode := cdr[0]
	if listNode.Kind != core.KindArray {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("car requires a list argument, got %v", listNode.Kind))
	}

	arr, ok := listNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("invalid array value: %T", listNode.Value))
	}

	if len(arr) == 0 {
		return nil, core.NewEvaluationError(listNode, "car: empty list")
	}

	firstElem, ok := arr[0].(*core.YispNode)
	if !ok {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("invalid element type: %T", arr[0]))
	}

	return firstElem, nil
}

// opCdr returns all but the first element of a list
func opCdr(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("cdr requires 1 argument, got %d", len(cdr)))
	}

	listNode := cdr[0]
	if listNode.Kind != core.KindArray {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("cdr requires a list argument, got %v", listNode.Kind))
	}

	arr, ok := listNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("invalid array value: %T", listNode.Value))
	}

	if len(arr) == 0 {
		return nil, core.NewEvaluationError(listNode, "cdr: empty list")
	}

	restElements := make([]any, len(arr)-1)
	copy(restElements, arr[1:])

	return &core.YispNode{
		Kind:  core.KindArray,
		Value: restElements,
	}, nil
}

// opCons constructs a new list by adding an element to the front of a list
func opCons(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("cons requires 2 arguments, got %d", len(cdr)))
	}

	elemNode := cdr[0]
	listNode := cdr[1]
	if listNode.Kind != core.KindArray {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("cons requires a list as the second argument, got %v", listNode.Kind))
	}

	arr, ok := listNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("invalid array value: %T", listNode.Value))
	}

	newArr := make([]any, len(arr)+1)
	newArr[0] = elemNode
	for i, elem := range arr {
		newArr[i+1] = elem
	}

	return &core.YispNode{
		Kind:  core.KindArray,
		Value: newArr,
	}, nil
}

func opFilter(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) < 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("filter requires at least 2 arguments, got %d", len(cdr)))
	}

	arrNode := cdr[0]
	fnNode := cdr[1]
	isDocumentRoot := true
	filtered := make([]any, 0)

	if arrNode.Kind != core.KindArray {
		return nil, core.NewEvaluationError(arrNode, fmt.Sprintf("filter requires an array as the first argument, got %v", arrNode.Kind))
	}

	arr, ok := arrNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(arrNode, fmt.Sprintf("invalid array value: %T", arrNode.Value))
	}

	for _, item := range arr {
		itemNode, ok := item.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(arrNode, fmt.Sprintf("invalid item type: %T", item))
		}
		if !itemNode.IsDocumentRoot {
			isDocumentRoot = false
		}
		args := []*core.YispNode{itemNode}
		rawResult, err := e.Apply(fnNode, args, env, mode)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(fnNode, "failed to evaluate filter function", err)
		}

		result, err := core.IsTruthy(rawResult)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(fnNode, "failed to evaluate filter function", err)
		}

		if result {
			filtered = append(filtered, itemNode)
		}
	}

	return &core.YispNode{
		Kind:           core.KindArray,
		Value:          filtered,
		IsDocumentRoot: isDocumentRoot,
	}, nil
}

func opFlatten(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	flattened := make([]any, 0)
	isDocumentRoot := true

	for _, node := range cdr {
		if !node.IsDocumentRoot {
			isDocumentRoot = false
		}

		if node.Kind == core.KindArray {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for flatten: %T", node))
			}
			for _, item := range arr {
				itemNode, ok := item.(*core.YispNode)
				if !ok {
					return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				flattened = append(flattened, itemNode)
			}
		} else {
			flattened = append(flattened, node)
		}
	}

	return &core.YispNode{
		Kind:           core.KindArray,
		Value:          flattened,
		IsDocumentRoot: isDocumentRoot,
	}, nil
}

func opMap(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) < 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("map requires more than 1 argument, got %d", len(cdr)))
	}

	arrNode := cdr[0]
	fnNode := cdr[1]

	if arrNode.Kind != core.KindArray {
		return nil, core.NewEvaluationError(arrNode, fmt.Sprintf("map requires an array as the first argument, got %v", arrNode.Kind))
	}

	arr, ok := arrNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(arrNode, fmt.Sprintf("invalid array value: %T", arrNode.Value))
	}

	results := make([]any, len(arr))

	for i, item := range arr {
		itemNode, ok := item.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(arrNode, fmt.Sprintf("invalid item type: %T", item))
		}

		args := []*core.YispNode{itemNode}
		result, err := e.Apply(fnNode, args, env, mode)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(fnNode, "failed to evaluate map function", err)
		}

		results[i] = result
	}

	return &core.YispNode{
		Kind:           core.KindArray,
		Attr:           arrNode.Attr,
		Value:          results,
		IsDocumentRoot: arrNode.IsDocumentRoot,
	}, nil

}

// opReduce applies a function cumulatively to the elements of a list, reducing it to a single value
func opReduce(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) < 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("reduce requires at least 2 arguments, got %d", len(cdr)))
	}

	listNode := cdr[0]
	if listNode.Kind != core.KindArray {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("reduce requires an array as the first argument, got %v", listNode.Kind))
	}

	fnNode := cdr[1]

	arr, ok := listNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(listNode, fmt.Sprintf("invalid array value: %T", listNode.Value))
	}

	if len(arr) == 0 {
		return nil, core.NewEvaluationError(listNode, "reduce: empty list")
	}

	accumulator := arr[0] // Start with the first element
	for _, item := range arr[1:] {
		itemNode, ok := item.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(listNode, fmt.Sprintf("invalid item type: %T", item))
		}
		result, err := e.Apply(fnNode, []*core.YispNode{accumulator.(*core.YispNode), itemNode}, env, mode)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(fnNode, "failed to evaluate reduce function", err)
		}
		accumulator = result
	}

	return accumulator.(*core.YispNode), nil
}

// opIota generates a list of integers from 0 to n-1
func opIota(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) < 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("iota requires at least 1 argument, got %d", len(cdr)))
	}

	nNode := cdr[0]
	n, ok := nNode.Value.(int)
	if !ok {
		return nil, core.NewEvaluationError(nNode, fmt.Sprintf("iota requires an integer argument, got %T", nNode.Value))
	}

	if n < 0 {
		return nil, core.NewEvaluationError(nNode, "iota requires a non-negative integer")
	}

	start := 0
	if len(cdr) >= 2 {
		startNode := cdr[1]
		start, ok = startNode.Value.(int)
		if !ok {
			return nil, core.NewEvaluationError(startNode, fmt.Sprintf("iota requires an integer start argument, got %T", startNode.Value))
		}
		if start < 0 {
			return nil, core.NewEvaluationError(startNode, "iota requires a non-negative start integer")
		}
	}

	result := make([]any, n)
	for i := 0; i < n; i++ {
		result[i] = &core.YispNode{
			Kind:  core.KindInt,
			Value: start + i,
		}
	}
	return &core.YispNode{
		Kind:  core.KindArray,
		Value: result,
	}, nil
}
