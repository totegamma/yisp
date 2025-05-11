package yisp

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
)

// OperatorFunc is a function that implements a Yisp operator
type OperatorFunc func([]*YispNode, *Env, EvalMode) (*YispNode, error)

// operators is a map of operator names to their implementations
var operators = make(map[string]OperatorFunc)

// init initializes the operators map
func init() {
	operators["concat"] = opConcat
	operators["+"] = opAdd
	operators["-"] = opSubtract
	operators["*"] = opMultiply
	operators["/"] = opDivide
	operators["if"] = opIf
	operators["=="] = opEqual
	operators["!="] = opNotEqual
	operators["<"] = opLessThan
	operators["<="] = opLessThanOrEqual
	operators[">"] = opGreaterThan
	operators[">="] = opGreaterThanOrEqual
	operators["car"] = opCar
	operators["cdr"] = opCdr
	operators["cons"] = opCons
	operators["discard"] = opDiscard
	operators["progn"] = opProgn
	operators["include"] = opInclude
	operators["import"] = opImport
	operators["lambda"] = opLambda
	operators["cmd"] = opCmd
	operators["getmap"] = opGetMap
}

// Call dispatches to the appropriate operator function based on the operator name
func Call(op string, cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if fn, ok := operators[op]; ok {
		return fn(cdr, env, mode)
	}

	JsonPrint("env", env)
	return nil, NewEvaluationError(nil, fmt.Sprintf("unknown function name: %s", op))
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

	cond, err := isTruthy(condNode)
	if err != nil {
		return nil, err
	}

	if cond {
		return Eval(cdr[1], env, mode)
	}
	return Eval(cdr[2], env, mode)
}

// opEqual checks if two values are equal
func opEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareValues(cdr, env, mode, "==", true)
}

// opNotEqual checks if two values are not equal
func opNotEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareValues(cdr, env, mode, "!=", false)
}

// opLessThan checks if the first number is less than the second
func opLessThan(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, env, mode, "<", func(a, b float64) bool { return a < b })
}

// opLessThanOrEqual checks if the first number is less than or equal to the second
func opLessThanOrEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, env, mode, "<=", func(a, b float64) bool { return a <= b })
}

// opGreaterThan checks if the first number is greater than the second
func opGreaterThan(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, env, mode, ">", func(a, b float64) bool { return a > b })
}

// opGreaterThanOrEqual checks if the first number is greater than or equal to the second
func opGreaterThanOrEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, env, mode, ">=", func(a, b float64) bool { return a >= b })
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
		_, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate discard argument: %s", err))
		}
	}

	return nil, nil
}

func opProgn(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	var result *YispNode
	var err error
	for _, node := range cdr {
		result, err = Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate progn argument: %s", err))
		}
	}

	return result, nil
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
		results[i], err = evaluateYispFile(path, env.CreateChild())
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
		_, err = evaluateYispFile(path, newEnv)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to include file: %s", err))
		}

		env.AddModule(name, newEnv)
	}

	return &YispNode{
		Kind: KindNull,
	}, nil
}

func opGetMap(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("map requires 1 argument, got %d", len(cdr)))
	}

	mapValue, err := EvalAndCastAny[map[string]any](cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate map argument: %s", err))
	}

	keyValue, err := EvalAndCastAny[string](cdr[1], env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("failed to evaluate key argument: %s", err))
	}

	value, ok := mapValue[keyValue]
	if !ok {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("key %s not found in map", keyValue))
	}

	valueNode, ok := value.(*YispNode)
	if !ok {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("invalid value type: %T", value))
	}

	return valueNode, nil
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
		Clojure: env.Clone(),
	}

	return &YispNode{
		Kind:  KindLambda,
		Value: lambda,
	}, nil
}

func opCmd(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	if allowCmd != true {
		return nil, NewEvaluationError(nil, "cmdline operator is not allowed. Set --allow-cmd to enable it.")
	}

	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("cmdline requires 1 argument, got %d", len(cdr)))
	}

	props := cdr[0]
	if props.Kind != KindMap {
		return nil, NewEvaluationError(props, fmt.Sprintf("cmdline requires a map argument, got %v", props.Kind))
	}

	propsMap, ok := props.Value.(map[string]any)
	if !ok {
		return nil, NewEvaluationError(props, fmt.Sprintf("invalid map type: %T", props.Value))
	}

	cmdAny, ok := propsMap["cmd"]
	if !ok {
		return nil, NewEvaluationError(props, "cmdline requires a 'cmd' key in the map")
	}

	cmdStr, err := EvalAndCastAny[string](cmdAny, env, mode)
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate cmd argument: %s", err))
	}

	args := make([]string, 0)

	argsAny, ok := propsMap["args"]
	if ok {
		argsNode, ok := argsAny.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(props, fmt.Sprintf("invalid args type: %T", argsAny))
		}

		if argsNode.Kind != KindArray {
			return nil, NewEvaluationError(argsNode, fmt.Sprintf("args must be an array, got %v", argsNode.Kind))
		}

		arr, ok := argsNode.Value.([]any)
		if !ok {
			return nil, NewEvaluationError(argsNode, fmt.Sprintf("invalid array value: %T", argsNode.Value))
		}

		for _, item := range arr {
			arg, err := EvalAndCastAny[string](item, env, mode)
			if err != nil {
				return nil, NewEvaluationError(argsNode, fmt.Sprintf("invalid arg value: %s", err))
			}
			args = append(args, arg)
		}
	}

	asString := false
	asStringAny, ok := propsMap["asString"]
	if ok {
		asString, err = EvalAndCastAny[bool](asStringAny, env, mode)
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := exec.Command(cmdStr, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	errorOutput := stderr.String()
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("command execution error: %s", errorOutput))
	}

	if asString {
		return &YispNode{
			Kind:  KindString,
			Value: stdout.String(),
		}, nil
	} else {

		result, err := evaluateYisp(stdout, env, cdr[0].File)
		if err != nil {
			return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate command output: %s", err))
		}

		return result, nil
	}
}
