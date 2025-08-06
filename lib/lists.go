package lib

import (
	"fmt"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("lists", "car", opCar)
	register("lists", "cdr", opCdr)
	register("lists", "cons", opCons)
	register("lists", "map", opMap)
	register("lists", "flatten", opFlatten)
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

func opMap(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) < 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("map requires more than 1 argument, got %d", len(cdr)))
	}

	fnNode := cdr[0]

	isDocumentRoot := true
	argList := make([][]any, len(cdr)-1)
	for i, node := range cdr[1:] {

		if !node.IsDocumentRoot {
			isDocumentRoot = false
		}

		if node.Kind != core.KindArray {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("map requires an array argument, got %v", node.Kind))
		}

		arg, ok := node.Value.([]any)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type: %T", node.Value))
		}
		yispList := make([]any, len(arg))
		for j, item := range arg {
			itemNode, ok := item.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
			}
			itemNode.Tag = "!quote"
			yispList[j] = itemNode
		}

		if i > 0 {
			if len(yispList) != len(argList[0]) {
				return nil, core.NewEvaluationError(node, "map requires all arguments to have the same length")
			}
		}

		argList[i] = yispList
	}

	results := make([]any, len(argList[0]))
	for i := range len(argList[0]) {
		args := []*core.YispNode{}
		for j := range argList {
			node, ok := argList[j][i].(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(fnNode, fmt.Sprintf("invalid item type: %T", argList[j][i]))
			}
			args = append(args, node)
		}
		result, err := e.Apply(fnNode, args, env, mode)

		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(fnNode, "failed to evaluate map argument", err)
		}
		results[i] = result
	}

	return &core.YispNode{
		Kind:           core.KindArray,
		Value:          results,
		Attr:           fnNode.Attr,
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
