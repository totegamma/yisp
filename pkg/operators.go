package yisp

import (
	"fmt"
	"path/filepath"
)

// Call dispatches to the appropriate operator function based on the operator name
func Call(op string, cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	switch op {
	case "concat":
		return opConcat(cdr, env, mode)
	case "+":
		return opAdd(cdr, env, mode)
	case "-":
		return opSubtract(cdr, env, mode)
	case "*":
		return opMultiply(cdr, env, mode)
	case "/":
		return opDivide(cdr, env, mode)
	case "if":
		return opIf(cdr, env, mode)
	case "==":
		return opEqual(cdr, env, mode)
	case "!=":
		return opNotEqual(cdr, env, mode)
	case "<":
		return opLessThan(cdr, env, mode)
	case "<=":
		return opLessThanOrEqual(cdr, env, mode)
	case ">":
		return opGreaterThan(cdr, env, mode)
	case ">=":
		return opGreaterThanOrEqual(cdr, env, mode)
	case "car":
		return opCar(cdr, env, mode)
	case "cdr":
		return opCdr(cdr, env, mode)
	case "cons":
		return opCons(cdr, env, mode)
	case "discard":
		return opDiscard(cdr, env, mode)
	case "include":
		return opInclude(cdr, env, mode)
	case "import":
		return opImport(cdr, env, mode)
	case "lambda":
		return opLambda(cdr, env, mode)
	default:
		JsonPrint("env", env)
		return nil, NewEvaluationError(nil, fmt.Sprintf("unknown function name: %s", op))
	}
}

// opConcat concatenates strings
func opConcat(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	var result string
	for _, node := range cdr {
		val, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate argument: %s", err))
		}
		str, ok := val.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for concat: %T", val))
		}
		result += str
	}

	return &YispNode{
		Kind:  KindString,
		Value: result,
	}, nil
}

// opAdd adds numbers
func opAdd(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	sum := 0
	for _, node := range cdr {
		val, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate argument: %s", err))
		}
		num, ok := val.Value.(int)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for +: %T", val))
		}
		sum += num
	}
	return &YispNode{
		Kind:  KindInt,
		Value: sum,
	}, nil
}

// opSubtract subtracts numbers
func opSubtract(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) == 0 {
		return &YispNode{
			Kind:  KindInt,
			Value: 0,
		}, nil
	}
	firstNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate first argument: %s", err))
	}
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for -: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		evaluated, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate argument: %s", err))
		}
		val, ok := evaluated.Value.(int)
		if !ok {
			return nil, NewEvaluationError(evaluated, fmt.Sprintf("invalid argument type for -: %T", evaluated))
		}
		baseNum -= val
	}
	return &YispNode{
		Kind:  KindInt,
		Value: baseNum,
	}, nil
}

// opMultiply multiplies numbers
func opMultiply(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	product := 1
	for _, node := range cdr {
		val, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate argument: %s", err))
		}
		num, ok := val.Value.(int)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for *: %T", val))
		}
		product *= num
	}
	return &YispNode{
		Kind:  KindInt,
		Value: product,
	}, nil
}

// opDivide divides numbers
func opDivide(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) == 0 {
		return &YispNode{
			Kind:  KindInt,
			Value: 0,
		}, nil
	}
	firstNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate first argument: %s", err))
	}
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for /: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		evaluated, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate argument: %s", err))
		}
		val, ok := evaluated.Value.(int)
		if !ok {
			return nil, NewEvaluationError(evaluated, fmt.Sprintf("invalid argument type for /: %T", evaluated))
		}
		if val == 0 {
			return nil, NewEvaluationError(evaluated, "division by zero")
		}
		baseNum /= val
	}
	return &YispNode{
		Kind:  KindInt,
		Value: baseNum,
	}, nil
}

// opIf implements conditional branching
func opIf(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 3 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("if requires 3 arguments, got %d", len(cdr)))
	}

	condNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate condition: %s", err))
	}

	cond := false
	switch condNode.Value.(type) {
	case bool:
		cond, _ = condNode.Value.(bool)
	case int:
		condInt, _ := condNode.Value.(int)
		cond = condInt != 0
	case float64:
		condFloat, _ := condNode.Value.(float64)
		cond = condFloat != 0.0
	case string:
		condStr, _ := condNode.Value.(string)
		cond = condStr != ""
	case []any:
		condArr, _ := condNode.Value.([]any)
		cond = len(condArr) != 0
	case map[string]any:
		condMap, _ := condNode.Value.(map[string]any)
		cond = len(condMap) != 0
	case nil:
		cond = false
	default:
		return nil, NewEvaluationError(condNode, fmt.Sprintf("invalid condition type: %T", condNode.Value))
	}

	if cond {
		return Eval(cdr[1], env, mode)
	}
	return Eval(cdr[2], env, mode)
}

// opEqual checks if two integers are equal
func opEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareInts(cdr, env, mode, "==", func(a, b int) bool { return a == b })
}

// opNotEqual checks if two integers are not equal
func opNotEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareInts(cdr, env, mode, "!=", func(a, b int) bool { return a != b })
}

// opLessThan checks if the first integer is less than the second
func opLessThan(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareInts(cdr, env, mode, "<", func(a, b int) bool { return a < b })
}

// opLessThanOrEqual checks if the first integer is less than or equal to the second
func opLessThanOrEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareInts(cdr, env, mode, "<=", func(a, b int) bool { return a <= b })
}

// opGreaterThan checks if the first integer is greater than the second
func opGreaterThan(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareInts(cdr, env, mode, ">", func(a, b int) bool { return a > b })
}

// opGreaterThanOrEqual checks if the first integer is greater than or equal to the second
func opGreaterThanOrEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareInts(cdr, env, mode, ">=", func(a, b int) bool { return a >= b })
}

// opCar returns the first element of a list
func opCar(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("car requires 1 argument, got %d", len(cdr)))
	}

	listNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate car argument: %s", err))
	}

	if listNode.Kind != KindArray {
		return nil, NewEvaluationError(listNode, fmt.Sprintf("car requires a list argument, got %v", listNode.Kind))
	}

	arr, ok := listNode.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(listNode, fmt.Sprintf("invalid array value: %T", listNode.Value))
	}

	if len(arr) == 0 {
		return nil, NewEvaluationError(listNode, "car: empty list")
	}

	firstElem, ok := arr[0].(*YispNode)
	if !ok {
		return nil, NewEvaluationError(listNode, fmt.Sprintf("invalid element type: %T", arr[0]))
	}

	return firstElem, nil
}

// opCdr returns all but the first element of a list
func opCdr(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("cdr requires 1 argument, got %d", len(cdr)))
	}

	listNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate cdr argument: %s", err))
	}

	if listNode.Kind != KindArray {
		return nil, NewEvaluationError(listNode, fmt.Sprintf("cdr requires a list argument, got %v", listNode.Kind))
	}

	arr, ok := listNode.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(listNode, fmt.Sprintf("invalid array value: %T", listNode.Value))
	}

	if len(arr) == 0 {
		return nil, NewEvaluationError(listNode, "cdr: empty list")
	}

	restElements := make([]any, len(arr)-1)
	for i, elem := range arr[1:] {
		restElements[i] = elem
	}

	return &YispNode{
		Kind:  KindArray,
		Value: restElements,
	}, nil
}

// opCons constructs a new list by adding an element to the front of a list
func opCons(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("cons requires 2 arguments, got %d", len(cdr)))
	}

	elemNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate cons first argument: %s", err))
	}

	listNode, err := Eval(cdr[1], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("failed to evaluate cons second argument: %s", err))
	}

	if listNode.Kind != KindArray {
		return nil, NewEvaluationError(listNode, fmt.Sprintf("cons requires a list as the second argument, got %v", listNode.Kind))
	}

	arr, ok := listNode.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(listNode, fmt.Sprintf("invalid array value: %T", listNode.Value))
	}

	newArr := make([]any, len(arr)+1)
	newArr[0] = elemNode
	for i, elem := range arr {
		newArr[i+1] = elem
	}

	return &YispNode{
		Kind:  KindArray,
		Value: newArr,
	}, nil
}

// opDiscard evaluates all arguments and returns nil
func opDiscard(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	for _, node := range cdr {
		Eval(node, env, mode)
	}

	return nil, nil
}

// opInclude includes files
func opInclude(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	results := make([]any, len(cdr))
	for i, node := range cdr {
		relpath, ok := node.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid path type: %T", node.Value))
		}

		baseDir := filepath.Dir(node.File)
		joinedPath := filepath.Join(baseDir, relpath)
		path := filepath.Clean(joinedPath)

		var err error
		results[i], err = evaluateYisp(path, env.CreateChild())
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to include file: %s", err))
		}
	}

	return &YispNode{
		Kind:  KindArray,
		Value: results,
		Tag:   "!expand",
	}, nil
}

// opImport imports modules
func opImport(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	for _, node := range cdr {
		baseDir := filepath.Dir(node.File)

		tuple, ok := node.Value.([]any)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid tuple type: %T", node.Value))
		}

		if len(tuple) != 2 {
			return nil, NewEvaluationError(node, fmt.Sprintf("import requires 2 arguments, got %d", len(tuple)))
		}

		nameNode, ok := tuple[0].(*YispNode)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid name type: %T", tuple[0]))
		}

		name, ok := nameNode.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid name type: %T", nameNode.Value))
		}

		relpathNode, ok := tuple[1].(*YispNode)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid path type: %T", tuple[1]))
		}

		relpath, ok := relpathNode.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid path type: %T", relpathNode.Value))
		}

		joinedPath := filepath.Join(baseDir, relpath)
		path := filepath.Clean(joinedPath)

		newEnv := NewEnv()

		var err error
		_, err = evaluateYisp(path, newEnv)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to include file: %s", err))
		}

		env.AddModule(name, newEnv)
	}

	return &YispNode{
		Kind: KindNull,
	}, nil
}

// opLambda creates a lambda function
func opLambda(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("lambda requires 2 arguments, got %d", len(cdr)))
	}

	paramsNode := cdr[0]
	bodyNode := cdr[1]

	params := make([]string, 0)
	for _, item := range paramsNode.Value.([]any) {
		paramNode, ok := item.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(nil, fmt.Sprintf("invalid param type: %T", item))
		}
		param, ok := paramNode.Value.(string)
		if !ok {
			return nil, NewEvaluationError(nil, fmt.Sprintf("invalid param value: %T", paramNode.Value))
		}
		params = append(params, param)
	}

	lambda := &Lambda{
		Params:  params,
		Body:    bodyNode,
		Clojure: env,
	}

	return &YispNode{
		Kind:  KindLambda,
		Value: lambda,
	}, nil
}

// compareInts compares two integers using the provided comparison function
func compareInts(cdr []*YispNode, env *Env, mode EvalMode, opName string, cmp func(int, int) bool) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("%s requires 2 arguments, got %d", opName, len(cdr)))
	}
	firstNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate first argument: %s", err))
	}
	firstNum, ok := firstNode.Value.(int)
	if !ok {
		// Attempt to convert float to int if applicable, or handle other types
		if firstFloat, isFloat := firstNode.Value.(float64); isFloat {
			firstNum = int(firstFloat) // Note: This truncates. Decide if this is the desired behavior.
		} else {
			return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid first argument type for %s: %T (value: %v)", opName, firstNode.Value, firstNode.Value))
		}
	}

	secondNode, err := Eval(cdr[1], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("failed to evaluate second argument: %s", err))
	}
	secondNum, ok := secondNode.Value.(int)
	if !ok {
		if secondFloat, isFloat := secondNode.Value.(float64); isFloat {
			secondNum = int(secondFloat) // Note: This truncates.
		} else {
			return nil, NewEvaluationError(secondNode, fmt.Sprintf("invalid second argument type for %s: %T (value: %v)", opName, secondNode.Value, secondNode.Value))
		}
	}

	return &YispNode{
		Kind:  KindBool,
		Value: cmp(firstNum, secondNum),
	}, nil
}
