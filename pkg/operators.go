package yisp

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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
	operators["and"] = opAnd
	operators["or"] = opOr
	operators["not"] = opNot
	operators["car"] = opCar
	operators["cdr"] = opCdr
	operators["cons"] = opCons
	operators["discard"] = opDiscard
	operators["progn"] = opProgn
	operators["include"] = opInclude
	operators["import"] = opImport
	operators["lambda"] = opLambda
	operators["cmd"] = opCmd
	operators["mapping-get"] = opMappingGet
	operators["merge"] = opMerge
	operators["map"] = opMap
	operators["flatten"] = opFlatten
	operators["read-files"] = opReadFiles
	operators["to-entries"] = opToEntries
	operators["from-entries"] = opFromEntries
	operators["to-yaml"] = opToYaml
	operators["sha256"] = opSha256
	operators["schema"] = opSchema
	operators["go-run"] = opGoRun
	operators["pipeline"] = opPipeline
	operators["format"] = opFormat
	operators["k8s-patch"] = opPatch
	operators["as-document-root"] = opAsDocumentRoot
}

// Call dispatches to the appropriate operator function based on the operator name
func Call(car *YispNode, cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	op, ok := car.Value.(string)
	if !ok {
		return nil, NewEvaluationError(car, fmt.Sprintf("invalid car value: %T", car.Value))
	}

	if fn, ok := operators[op]; ok {
		return fn(cdr, env, mode)
	}

	return nil, NewEvaluationError(car, fmt.Sprintf("unknown function name: %s", op))
}

// opConcat concatenates strings
func opConcat(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	var result string
	for _, node := range cdr {
		val, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
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
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
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
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate first argument"), err)
	}
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for -: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		evaluated, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
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
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
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
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate first argument"), err)
	}
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for /: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		evaluated, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
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
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate condition"), err)
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
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate car argument"), err)
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
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate cdr argument"), err)
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
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate cons first argument"), err)
	}

	listNode, err := Eval(cdr[1], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[1], fmt.Sprintf("failed to evaluate cons second argument"), err)
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
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate discard argument"), err)
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
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate progn argument"), err)
		}
	}

	return result, nil
}

// opInclude includes files
func opInclude(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	results := make([]any, 0)
	for _, node := range cdr {
		relpath, ok := node.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid path type: %T", node.Value))
		}

		var err error
		evaluated, err := evaluateYispFile(relpath, node.Pos.File, NewEnv())
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to include file"), err)
		}

		if evaluated.Kind != KindArray {
			return nil, NewEvaluationError(node, fmt.Sprintf("include requires an array result, got %v", evaluated.Kind))
		}
		arr, ok := evaluated.Value.([]any)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid array value: %T", evaluated.Value))
		}

		for _, item := range arr {
			itemNode, ok := item.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
			}
			itemNode.Tag = "!quote"
			results = append(results, itemNode)
		}
	}

	return &YispNode{
		Kind:           KindArray,
		Value:          results,
		IsDocumentRoot: true,
	}, nil
}

// opImport imports modules
func opImport(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	for _, node := range cdr {

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

		newEnv := NewEnv()

		var err error
		_, err = evaluateYispFile(relpath, node.Pos.File, newEnv)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to include file"), err)
		}

		env.Root().Set(name, &YispNode{
			Kind:  KindMap,
			Value: newEnv.Vars,
		})
	}

	return &YispNode{
		Kind: KindNull,
	}, nil
}

func opMap(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) < 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("map requires more than 1 argument, got %d", len(cdr)))
	}

	fnNode := cdr[0]

	isDocumentRoot := true
	argList := make([][]any, len(cdr)-1)
	for i, node := range cdr[1:] {
		argEvaluated, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate map argument"), err)
		}

		if !argEvaluated.IsDocumentRoot {
			isDocumentRoot = false
		}

		if argEvaluated.Kind != KindArray {
			return nil, NewEvaluationError(node, fmt.Sprintf("map requires an array argument, got %v", argEvaluated.Kind))
		}

		arg, ok := argEvaluated.Value.([]any)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type: %T", node.Value))
		}
		yispList := make([]any, len(arg))
		for j, item := range arg {
			itemNode, ok := item.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
			}
			itemNode.Tag = "!quote"
			yispList[j] = itemNode
		}

		if i > 0 {
			if len(yispList) != len(argList[0]) {
				return nil, NewEvaluationError(node, fmt.Sprintf("map requires all arguments to have the same length"))
			}
		}

		argList[i] = yispList
	}

	results := make([]any, len(argList[0]))
	for i := range len(argList[0]) {
		code := []any{fnNode}
		for j := range argList {
			code = append(code, argList[j][i])
		}
		result, err := Eval(&YispNode{
			Kind:  KindArray,
			Value: code,
			Pos:   fnNode.Pos,
		}, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(fnNode, fmt.Sprintf("failed to evaluate map argument"), err)
		}
		results[i] = result
	}

	return &YispNode{
		Kind:           KindArray,
		Value:          results,
		Pos:            fnNode.Pos,
		IsDocumentRoot: isDocumentRoot,
	}, nil
}

func opMappingGet(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("map requires 1 argument, got %d", len(cdr)))
	}

	mapValue, err := EvalAndCastAny[*YispMap](cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate map argument"), err)
	}

	keyValue, err := EvalAndCastAny[string](cdr[1], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[1], fmt.Sprintf("failed to evaluate key argument"), err)
	}

	value, ok := mapValue.Get(keyValue)
	if !ok {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("key %s not found in map", keyValue))
	}

	valueNode, ok := value.(*YispNode)
	if !ok {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("invalid value type: %T", value))
	}

	return valueNode, nil
}

func opMerge(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	result := &YispNode{
		Kind: KindNull,
	}
	for _, node := range cdr {
		value, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate map argument"), err)
		}

		result, err = DeepMergeYispNode(result, value, value.Type)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to merge map"), err)
		}
	}

	return result, nil
}

// opLambda creates a lambda function
func opLambda(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("lambda requires 2 arguments, got %d", len(cdr)))
	}

	paramsNode := cdr[0]
	bodyNode := cdr[1]

	params := make([]TypedSymbol, 0)
	for _, item := range paramsNode.Value.([]any) {
		paramNode, ok := item.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(nil, fmt.Sprintf("invalid param type: %T", item))
		}
		param, ok := paramNode.Value.(string)
		if !ok {
			return nil, NewEvaluationError(nil, fmt.Sprintf("invalid param value: %T", paramNode.Value))
		}

		var schema *Schema
		tag := paramNode.Tag
		typeName := strings.TrimPrefix(tag, "!")
		if typeName != "" && !strings.HasPrefix(typeName, "!") {
			typeNode, ok := env.Get(typeName)
			if !ok {
				return nil, NewEvaluationError(nil, fmt.Sprintf("undefined type: %s", typeName))
			}
			if typeNode.Kind != KindType {
				return nil, NewEvaluationError(nil, fmt.Sprintf("%s is not a type. actual: %s", typeName, typeNode.Kind))
			}
			schema, ok = typeNode.Value.(*Schema)
			if !ok {
				return nil, NewEvaluationError(nil, fmt.Sprintf("invalid type value: %T", typeNode.Value))
			}
		}

		params = append(params, TypedSymbol{
			Name:   param,
			Schema: schema,
		})
	}

	var schema *Schema
	tag := paramsNode.Tag
	typeName := strings.TrimPrefix(tag, "!")
	if typeName != "" && !strings.HasPrefix(typeName, "!") {
		typeNode, ok := env.Get(typeName)
		if !ok {
			return nil, NewEvaluationError(nil, fmt.Sprintf("undefined type: %s", typeName))
		}
		if typeNode.Kind != KindType {
			return nil, NewEvaluationError(nil, fmt.Sprintf("%s is not a type. actual: %s", typeName, typeNode.Kind))
		}
		schema, ok = typeNode.Value.(*Schema)
		if !ok {
			return nil, NewEvaluationError(nil, fmt.Sprintf("invalid type value: %T", typeNode.Value))
		}
	}

	lambda := &Lambda{
		Arguments: params,
		Returns:   schema,
		Body:      bodyNode,
		Clojure:   env.Clone(),
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

	propsMap, ok := props.Value.(*YispMap)
	if !ok {
		return nil, NewEvaluationError(props, fmt.Sprintf("invalid map type: %T", props.Value))
	}

	cmdAny, ok := propsMap.Get("cmd")
	if !ok {
		return nil, NewEvaluationError(props, "cmdline requires a 'cmd' key in the map")
	}

	cmdStr, err := EvalAndCastAny[string](cmdAny, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate cmd argument"), err)
	}
	cmd := exec.Command(cmdStr)

	argsAny, ok := propsMap.Get("args")
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
				return nil, NewEvaluationErrorWithParent(argsNode, fmt.Sprintf("failed to evaluate arg"), err)
			}
			cmd.Args = append(cmd.Args, arg)
		}
	}

	stdinAny, ok := propsMap.Get("stdin")
	if ok {
		stdinNode, ok := stdinAny.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(props, fmt.Sprintf("invalid stdin type: %T", stdinAny))
		}

		if stdinNode.Kind != KindString {
			return nil, NewEvaluationError(stdinNode, fmt.Sprintf("stdin must be a string, got %v", stdinNode.Kind))
		}

		str, ok := stdinNode.Value.(string)
		if !ok {
			return nil, NewEvaluationError(stdinNode, fmt.Sprintf("invalid string value: %T", stdinNode.Value))
		}
		stdin := bytes.NewBufferString(str)
		cmd.Stdin = stdin
	}

	asString := false
	asStringAny, ok := propsMap.Get("asString")
	if ok {
		asString, err = EvalAndCastAny[bool](asStringAny, env, mode)
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
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

		result, err := evaluateYisp(stdout, env, cdr[0].Pos.File)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate command output"), err)
		}

		return result, nil
	}
}

func opGoRun(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("gorun requires 1 argument, got %d", len(cdr)))
	}
	props := cdr[0]
	if props.Kind != KindMap {
		return nil, NewEvaluationError(props, fmt.Sprintf("gorun requires a map argument, got %v", props.Kind))
	}
	propsMap, ok := props.Value.(*YispMap)
	if !ok {
		return nil, NewEvaluationError(props, fmt.Sprintf("invalid map type: %T", props.Value))
	}

	pkgAny, ok := propsMap.Get("pkg")
	if !ok {
		return nil, NewEvaluationError(props, "gorun requires a 'pkg' key in the map")
	}
	pkgStr, err := EvalAndCastAny[string](pkgAny, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate pkg argument"), err)
	}

	allowed := false
	for _, stmt := range allowedGoPkgs {
		regex := "^" + strings.ReplaceAll(regexp.QuoteMeta(stmt), "\\*", ".*") + "$"
		matched, err := regexp.MatchString(regex, pkgStr)
		if err != nil {
			return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to match package %s with regex %s: %v", pkgStr, regex, err))
		}
		if matched {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("package %s is not allowed. Run command below to allow it:\n\nyisp allow %s", pkgStr, pkgStr))
	}

	cmd := exec.Command("go", "run", pkgStr)

	argsAny, ok := propsMap.Get("args")
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
				return nil, NewEvaluationErrorWithParent(argsNode, fmt.Sprintf("failed to evaluate arg"), err)
			}
			cmd.Args = append(cmd.Args, arg)
		}
	}

	stdinAny, ok := propsMap.Get("stdin")
	if ok {
		str, err := EvalAndCastAny[string](stdinAny, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate stdin argument"), err)
		}
		stdin := bytes.NewBufferString(str)
		cmd.Stdin = stdin
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	errorOutput := stderr.String()
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("command execution error: %s", errorOutput))
	}

	asString := false
	asStringAny, ok := propsMap.Get("asString")
	if ok {
		asString, err = EvalAndCastAny[bool](asStringAny, env, mode)
	}

	if asString {
		return &YispNode{
			Kind:  KindString,
			Value: stdout.String(),
		}, nil
	} else {

		result, err := evaluateYisp(stdout, env, cdr[0].Pos.File)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate command output"), err)
		}

		return result, nil
	}
}

func opFlatten(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	flattened := make([]any, 0)
	isDocumentRoot := true

	for _, node := range cdr {
		val, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
		}

		if !val.IsDocumentRoot {
			isDocumentRoot = false
		}

		if val.Kind == KindArray {
			arr, ok := val.Value.([]any)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for flatten: %T", val))
			}
			for _, item := range arr {
				itemNode, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				flattened = append(flattened, itemNode)
			}
		} else {
			flattened = append(flattened, val)
		}
	}

	return &YispNode{
		Kind:           KindArray,
		Value:          flattened,
		IsDocumentRoot: isDocumentRoot,
	}, nil
}

func opReadFiles(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	result := make([]any, 0)

	for _, node := range cdr {
		val, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
		}
		str, ok := val.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for open: %T", val))
		}

		path := str
		if val.Pos.File != "" {
			path = filepath.Clean(filepath.Join(filepath.Dir(val.Pos.File), str))
		}

		files, err := filepath.Glob(path)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to glob path: %s", str))
		}

		for _, file := range files {

			filename := filepath.Base(file)
			body, err := os.ReadFile(file)
			if err != nil {
				return nil, NewEvaluationError(node, fmt.Sprintf("failed to read file: %s", file))
			}

			value := NewYispMap()
			value.Set("path", &YispNode{
				Kind:  KindString,
				Value: file,
			})
			value.Set("name", &YispNode{
				Kind:  KindString,
				Value: filename,
			})
			value.Set("body", &YispNode{
				Kind:  KindString,
				Value: string(body),
			})

			result = append(result, &YispNode{
				Kind:  KindMap,
				Value: value,
				Pos: Position{
					File:   file,
					Line:   node.Pos.Line,
					Column: node.Pos.Column,
				},
			})
		}
	}

	return &YispNode{
		Kind:  KindArray,
		Value: result,
		Pos:   cdr[0].Pos,
	}, nil
}

func opToEntries(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("toEntries requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	mapValue, err := EvalAndCastAny[*YispMap](node, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
	}
	result := make([]any, 0)
	for key, value := range mapValue.AllFromFront() {
		tuple := &YispNode{
			Kind: KindArray,
			Value: []any{
				&YispNode{
					Kind:  KindString,
					Value: key,
					Pos:   node.Pos,
				},
				value,
			},
		}
		result = append(result, tuple)
	}
	return &YispNode{
		Kind:  KindArray,
		Value: result,
		Pos:   node.Pos,
	}, nil
}

func opFromEntries(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	node := cdr[0]
	arr, err := EvalAndCastAny[[]any](node, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
	}

	result := NewYispMap()
	for _, item := range arr {
		tupleNode, ok := item.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid tuple type: %T", item))
		}

		tupleArr, ok := tupleNode.Value.([]any)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid tuple value: %T", tupleNode.Value))
		}

		if len(tupleArr) != 2 {
			return nil, NewEvaluationError(node, fmt.Sprintf("tuple must have exactly 2 elements"))
		}

		keyNode := tupleArr[0]
		valueNode := tupleArr[1]

		keyStr, err := EvalAndCastAny[string](keyNode, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate key"), err)
		}

		result.Set(keyStr, valueNode)
	}

	return &YispNode{
		Kind:  KindMap,
		Value: result,
		Pos:   node.Pos,
	}, nil
}

func opToYaml(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("toYaml requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	evaluated, err := Eval(node, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
	}

	yamlBytes, err := Render(evaluated)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to render yaml"), err)
	}
	yamlStr := string(yamlBytes)

	return &YispNode{
		Kind:  KindString,
		Value: yamlStr,
		Pos:   node.Pos,
	}, nil
}

// opAnd implements logical AND operation
func opAnd(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) == 0 {
		return &YispNode{
			Kind:  KindBool,
			Value: true,
		}, nil
	}

	for _, node := range cdr {
		evaluated, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
		}

		truthy, err := isTruthy(evaluated)
		if err != nil {
			return nil, err
		}

		if !truthy {
			return &YispNode{
				Kind:  KindBool,
				Value: false,
			}, nil
		}
	}

	return &YispNode{
		Kind:  KindBool,
		Value: true,
	}, nil
}

// opOr implements logical OR operation
func opOr(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) == 0 {
		return &YispNode{
			Kind:  KindBool,
			Value: false,
		}, nil
	}

	for _, node := range cdr {
		evaluated, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
		}

		truthy, err := isTruthy(evaluated)
		if err != nil {
			return nil, err
		}

		if truthy {
			return &YispNode{
				Kind:  KindBool,
				Value: true,
			}, nil
		}
	}

	return &YispNode{
		Kind:  KindBool,
		Value: false,
	}, nil
}

// opNot implements logical NOT operation
func opNot(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("not requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]
	evaluated, err := Eval(node, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
	}

	truthy, err := isTruthy(evaluated)
	if err != nil {
		return nil, err
	}

	return &YispNode{
		Kind:  KindBool,
		Value: !truthy,
	}, nil
}

func opSha256(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("sha256 requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	evaluated, err := Eval(node, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
	}

	if evaluated.Kind != KindString {
		return nil, NewEvaluationError(node, fmt.Sprintf("sha256 requires a string argument, got %v", evaluated.Kind))
	}

	str, ok := evaluated.Value.(string)
	if !ok {
		return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for sha256: %T", evaluated.Value))
	}

	hash := sha256.Sum256([]byte(str))

	return &YispNode{
		Kind:  KindString,
		Value: fmt.Sprintf("%x", hash),
		Pos:   node.Pos,
	}, nil
}

func opSchema(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("sha256 requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	evaluated, err := Eval(node, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
	}

	if evaluated.Kind != KindMap {
		return nil, NewEvaluationError(node, fmt.Sprintf("schema requires a map argument, got %v", evaluated.Kind))
	}

	rendered, err := ToNative(evaluated)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to render schema"), err)
	}
	schemaBytes, err := json.Marshal(rendered)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to marshal schema"), err)
	}

	var schema Schema
	err = json.Unmarshal(schemaBytes, &schema)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to unmarshal schema"), err)
	}

	return &YispNode{
		Kind:  KindType,
		Value: &schema,
		Pos:   node.Pos,
	}, nil

}

func opPipeline(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	value, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate pipeline"), err)
	}

	isDocumentRoot := value.IsDocumentRoot

	for _, fn := range cdr[1:] {
		value.Tag = "!quote"

		code := []any{
			fn,
			value,
		}

		value, err = Eval(&YispNode{
			Kind:  KindArray,
			Value: code,
			Pos:   fn.Pos,
		}, env, mode)

		if err != nil {
			return nil, NewEvaluationErrorWithParent(fn, fmt.Sprintf("failed to evaluate pipeline"), err)
		}
	}

	value.IsDocumentRoot = isDocumentRoot

	return value, nil
}

func opFormat(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	formatNode := cdr[0]
	argsNode := cdr[1:]

	formatStr, err := EvalAndCastAny[string](formatNode, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(formatNode, fmt.Sprintf("failed to evaluate format argument"), err)
	}

	args := make([]any, len(argsNode))
	for i, argNode := range argsNode {
		arg, err := Eval(argNode, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(argNode, fmt.Sprintf("failed to evaluate format argument"), err)
		}
		args[i] = arg.Value
	}

	return &YispNode{
		Kind:  KindString,
		Value: fmt.Sprintf(formatStr, args...),
		Pos:   formatNode.Pos,
	}, nil

}

func opPatch(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("patch requires 2 arguments, got %d", len(cdr)))
	}

	targetNodes := cdr[0]
	patchNodes := cdr[1]

	targets, err := Eval(targetNodes, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(targetNodes, fmt.Sprintf("failed to evaluate target"), err)
	}

	patchs, err := Eval(patchNodes, env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(patchNodes, fmt.Sprintf("failed to evaluate patch"), err)
	}

	if targets.Kind != KindArray || patchs.Kind != KindArray {
		return nil, NewEvaluationError(nil, "patch requires both target and patch to be maps")
	}

	targetArray, ok := targets.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(targetNodes, fmt.Sprintf("invalid target type: %T", targets.Value))
	}

	patchArray, ok := patchs.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(patchNodes, fmt.Sprintf("invalid patch type: %T", patchs.Value))
	}

	for _, patchAny := range patchArray {
		patchNode, ok := patchAny.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(patchNodes, fmt.Sprintf("invalid patch item type: %T", patchAny))
		}

		patchGVK, err := GetGVK(patchNode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(patchNode, fmt.Sprintf("failed to get GVK from patch"), err)
		}

		for i, targetAny := range targetArray {
			targetNode, ok := targetAny.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(targetNodes, fmt.Sprintf("invalid target item type: %T", targetAny))
			}

			targetGVK, err := GetGVK(targetNode)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(targetNode, fmt.Sprintf("failed to get GVK from target"), err)
			}

			k8sType, err := GetK8sSchema(targetGVK.Group, targetGVK.Version, targetGVK.Kind)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(targetNode, fmt.Sprintf("failed to get k8s schema for %s", targetGVK.String()), err)
			}

			if patchGVK.Equal(targetGVK) {

				targetArray[i], err = DeepMergeYispNode(targetNode, patchNode, k8sType)
				if err != nil {
					return nil, NewEvaluationErrorWithParent(patchNode, fmt.Sprintf("failed to apply patch"), err)
				}

			}
		}
	}

	return targets, nil
}

func opAsDocumentRoot(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	flattened := make([]any, 0)

	for _, node := range cdr {
		val, err := Eval(node, env, mode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
		}

		if val.Kind == KindArray {
			arr, ok := val.Value.([]any)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for flatten: %T", val))
			}
			for _, item := range arr {
				itemNode, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				flattened = append(flattened, itemNode)
			}
		} else {
			flattened = append(flattened, val)
		}
	}

	return &YispNode{
		Kind:           KindArray,
		Value:          flattened,
		IsDocumentRoot: true,
	}, nil
}
