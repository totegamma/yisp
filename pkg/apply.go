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

// Apply applies a function to arguments
func Apply(car *YispNode, cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	switch car.Kind {
	case KindLambda:
		lambda, ok := car.Value.(*Lambda)
		if !ok {
			return nil, NewEvaluationError(car, fmt.Sprintf("invalid lambda type: %T", car.Value))
		}

		newEnv := lambda.Clojure.CreateChild()
		for i, node := range cdr {
			if lambda.Arguments[i].Schema != nil {
				err := lambda.Arguments[i].Schema.Validate(node)
				if err != nil {
					return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("object does not satisfy type"), err)
				}
			}

			newEnv.Vars[lambda.Arguments[i].Name] = node
		}

		return Eval(lambda.Body, newEnv, mode)

	case KindString:
		op, ok := car.Value.(string)
		if !ok {
			return nil, NewEvaluationError(car, fmt.Sprintf("invalid car value: %T", car.Value))
		}

		fn, ok := operators[op]
		if !ok {
			return nil, NewEvaluationError(car, fmt.Sprintf("unknown function name: %s", op))
		}

		// Call the operator function with the arguments
		return fn(cdr, env.CreateChild(), mode)

	default:
		return nil, NewEvaluationError(car, fmt.Sprintf("cannot apply type %s", car.Kind))
	}
}

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
	operators["patch"] = opPatch
	operators["as-document-root"] = opAsDocumentRoot
	operators["assert-type"] = opAssertType
	operators["get-type"] = opGetType
	operators["typeof"] = opTypeOf
}

// opConcat concatenates strings
func opConcat(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	var result string
	for _, node := range cdr {
		str, ok := node.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for concat: %T", node))
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
		num, ok := node.Value.(int)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for +: %T", node))
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
	firstNode := cdr[0]
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for -: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		val, ok := node.Value.(int)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for -: %T", node))
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
		num, ok := node.Value.(int)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for *: %T", node))
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
	firstNode := cdr[0]
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for /: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		val, ok := node.Value.(int)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for /: %T", node))
		}
		if val == 0 {
			return nil, NewEvaluationError(node, "division by zero")
		}
		baseNum /= val
	}
	return &YispNode{
		Kind:  KindInt,
		Value: baseNum,
	}, nil
}

// opEqual checks if two values are equal
func opEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareValues(cdr, "==", true)
}

// opNotEqual checks if two values are not equal
func opNotEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareValues(cdr, "!=", false)
}

// opLessThan checks if the first number is less than the second
func opLessThan(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, "<", func(a, b float64) bool { return a < b })
}

// opLessThanOrEqual checks if the first number is less than or equal to the second
func opLessThanOrEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, "<=", func(a, b float64) bool { return a <= b })
}

// opGreaterThan checks if the first number is greater than the second
func opGreaterThan(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, ">", func(a, b float64) bool { return a > b })
}

// opGreaterThanOrEqual checks if the first number is greater than or equal to the second
func opGreaterThanOrEqual(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return compareNumbers(cdr, ">=", func(a, b float64) bool { return a >= b })
}

// opCar returns the first element of a list
func opCar(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("car requires 1 argument, got %d", len(cdr)))
	}

	listNode := cdr[0]
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

	listNode := cdr[0]
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

	elemNode := cdr[0]
	listNode := cdr[1]
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
	return &YispNode{
		Kind:  KindNull,
		Value: nil,
	}, nil
}

func opProgn(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	return cdr[len(cdr)-1], nil
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
		evaluated, err := evaluateYispFile(relpath, node.Attr.File, NewEnv())
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

func opMap(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) < 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("map requires more than 1 argument, got %d", len(cdr)))
	}

	fnNode := cdr[0]

	isDocumentRoot := true
	argList := make([][]any, len(cdr)-1)
	for i, node := range cdr[1:] {

		if !node.IsDocumentRoot {
			isDocumentRoot = false
		}

		if node.Kind != KindArray {
			return nil, NewEvaluationError(node, fmt.Sprintf("map requires an array argument, got %v", node.Kind))
		}

		arg, ok := node.Value.([]any)
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
		args := []*YispNode{}
		for j := range argList {
			node, ok := argList[j][i].(*YispNode)
			if !ok {
				return nil, NewEvaluationError(fnNode, fmt.Sprintf("invalid item type: %T", argList[j][i]))
			}
			args = append(args, node)
		}
		result, err := Apply(fnNode, args, env, mode)

		if err != nil {
			return nil, NewEvaluationErrorWithParent(fnNode, fmt.Sprintf("failed to evaluate map argument"), err)
		}
		results[i] = result
	}

	return &YispNode{
		Kind:           KindArray,
		Value:          results,
		Attr:           fnNode.Attr,
		IsDocumentRoot: isDocumentRoot,
	}, nil
}

func opMappingGet(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("map requires 1 argument, got %d", len(cdr)))
	}

	mapValue, ok := cdr[0].Value.(*YispMap)
	if !ok {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("mapping-get requires a map argument, got %v", cdr[0].Kind))
	}

	keyValue, ok := cdr[1].Value.(string)
	if !ok {
		return nil, NewEvaluationError(cdr[1], fmt.Sprintf("mapping-get requires a string key, got %v", cdr[1].Kind))
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
		var err error
		result, err = DeepMergeYispNode(result, node, node.Type)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to merge map"), err)
		}
	}

	return result, nil
}

func opCmd(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

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

	cmdNode, ok := cmdAny.(*YispNode)
	if !ok {
		return nil, NewEvaluationError(props, fmt.Sprintf("invalid cmd type: %T", cmdAny))
	}
	cmdStr, ok := cmdNode.Value.(string)
	if !ok {
		return nil, NewEvaluationError(cmdNode, fmt.Sprintf("invalid cmd value: %T", cmdNode.Value))
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
			node, ok := item.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(argsNode, fmt.Sprintf("invalid item type: %T", item))
			}

			arg, ok := node.Value.(string)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid arg type: %T", node.Value))
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
		asStringNode, ok := asStringAny.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(props, fmt.Sprintf("invalid asString type: %T", asStringAny))
		}
		asString, ok = asStringNode.Value.(bool)
		if !ok {
			return nil, NewEvaluationError(asStringNode, fmt.Sprintf("invalid asString value: %T", asStringNode.Value))
		}
	}

	if allowCmd != true {
		fmt.Fprintf(os.Stderr, "Going to run command: %v\n", cmd.Args)
		fmt.Fprintf(os.Stderr, "Press Enter to continue or Ctrl+C to cancel...\n")
		_, err := os.Stdin.Read(make([]byte, 1))
		if err != nil {
			return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to read input: %v", err))
		}
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
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

		result, err := evaluateYisp(stdout, env, cdr[0].Attr.File)
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

	pkgNode, ok := pkgAny.(*YispNode)
	if !ok {
		return nil, NewEvaluationError(props, fmt.Sprintf("invalid pkg type: %T", pkgAny))
	}
	pkgStr, ok := pkgNode.Value.(string)
	if !ok {
		return nil, NewEvaluationError(pkgNode, fmt.Sprintf("invalid pkg value: %T", pkgNode.Value))
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
			node, ok := item.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(argsNode, fmt.Sprintf("invalid item type: %T", item))
			}
			arg, ok := node.Value.(string)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid arg type: %T", node.Value))
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
		str, ok := stdinNode.Value.(string)
		if !ok {
			return nil, NewEvaluationError(stdinNode, fmt.Sprintf("invalid string value: %T", stdinNode.Value))
		}
		stdin := bytes.NewBufferString(str)
		cmd.Stdin = stdin
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	errorOutput := stderr.String()
	if err != nil {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("command execution error: %s", errorOutput))
	}

	asString := false
	asStringAny, ok := propsMap.Get("asString")
	if ok {
		asStringNode, ok := asStringAny.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(props, fmt.Sprintf("invalid asString type: %T", asStringAny))
		}
		asString, ok = asStringNode.Value.(bool)
		if !ok {
			return nil, NewEvaluationError(asStringNode, fmt.Sprintf("invalid asString value: %T", asStringNode.Value))
		}
	}

	if asString {
		return &YispNode{
			Kind:  KindString,
			Value: stdout.String(),
		}, nil
	} else {

		result, err := evaluateYisp(stdout, env, cdr[0].Attr.File)
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
		if !node.IsDocumentRoot {
			isDocumentRoot = false
		}

		if node.Kind == KindArray {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for flatten: %T", node))
			}
			for _, item := range arr {
				itemNode, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				flattened = append(flattened, itemNode)
			}
		} else {
			flattened = append(flattened, node)
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
		str, ok := node.Value.(string)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for open: %T", node))
		}

		path := str
		if node.Attr.File != "" {
			path = filepath.Clean(filepath.Join(filepath.Dir(node.Attr.File), str))
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
				Attr: Attribute{
					File:   file,
					Line:   node.Attr.Line,
					Column: node.Attr.Column,
				},
			})
		}
	}

	return &YispNode{
		Kind:  KindArray,
		Value: result,
		Attr:  cdr[0].Attr,
	}, nil
}

func opToEntries(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("toEntries requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	mapValue, ok := node.Value.(*YispMap)
	if !ok {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("toEntries requires a map argument, got %v", cdr[0].Kind))
	}
	result := make([]any, 0)
	for key, value := range mapValue.AllFromFront() {
		tuple := &YispNode{
			Kind: KindArray,
			Value: []any{
				&YispNode{
					Kind:  KindString,
					Value: key,
					Attr:  node.Attr,
				},
				value,
			},
		}
		result = append(result, tuple)
	}
	return &YispNode{
		Kind:  KindArray,
		Value: result,
		Attr:  node.Attr,
	}, nil
}

func opFromEntries(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	node := cdr[0]
	arr, ok := node.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(node, fmt.Sprintf("fromEntries requires an array argument, got %v", node.Kind))
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

		keyNode, ok := tupleArr[0].(*YispNode)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid key type: %T", tupleArr[0]))
		}
		valueNode := tupleArr[1]

		keyStr, ok := keyNode.Value.(string)
		if !ok {
			return nil, NewEvaluationError(keyNode, fmt.Sprintf("invalid key value: %T", keyNode.Value))
		}

		result.Set(keyStr, valueNode)
	}

	return &YispNode{
		Kind:  KindMap,
		Value: result,
		Attr:  node.Attr,
	}, nil
}

func opToYaml(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("toYaml requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	yamlBytes, err := Render(node)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to render yaml"), err)
	}
	yamlStr := string(yamlBytes)

	return &YispNode{
		Kind:  KindString,
		Value: yamlStr,
		Attr:  node.Attr,
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
		truthy, err := isTruthy(node)
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
		truthy, err := isTruthy(node)
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
	truthy, err := isTruthy(node)
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
	if node.Kind != KindString {
		return nil, NewEvaluationError(node, fmt.Sprintf("sha256 requires a string argument, got %v", node.Kind))
	}

	str, ok := node.Value.(string)
	if !ok {
		return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for sha256: %T", node.Value))
	}

	hash := sha256.Sum256([]byte(str))

	return &YispNode{
		Kind:  KindString,
		Value: fmt.Sprintf("%x", hash),
		Attr:  node.Attr,
	}, nil
}

func opSchema(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("sha256 requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	if node.Kind != KindMap {
		return nil, NewEvaluationError(node, fmt.Sprintf("schema requires a map argument, got %v", node.Kind))
	}

	rendered, err := ToNative(node)
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
		Attr:  node.Attr,
	}, nil

}

func opPipeline(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	value := cdr[0]
	isDocumentRoot := value.IsDocumentRoot

	for _, fn := range cdr[1:] {
		var err error
		value, err = Apply(fn, []*YispNode{value}, env, mode)
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

	formatStr, ok := formatNode.Value.(string)
	if !ok {
		return nil, NewEvaluationError(formatNode, fmt.Sprintf("format requires a string argument, got %v", formatNode.Kind))
	}

	args := make([]any, len(argsNode))
	for i, arg := range argsNode {
		args[i] = arg.Value
	}

	return &YispNode{
		Kind:  KindString,
		Value: fmt.Sprintf(formatStr, args...),
		Attr:  formatNode.Attr,
	}, nil

}

func opPatch(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("patch requires 2 arguments, got %d", len(cdr)))
	}

	targets := cdr[0]
	patches := cdr[1]

	if targets.Kind != KindArray || patches.Kind != KindArray {
		return nil, NewEvaluationError(nil, "patch requires both target and patch to be maps")
	}

	targetArray, ok := targets.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(targets, fmt.Sprintf("invalid target type: %T", targets.Value))
	}

	patchArray, ok := patches.Value.([]any)
	if !ok {
		return nil, NewEvaluationError(patches, fmt.Sprintf("invalid patch type: %T", patches.Value))
	}

	for _, patchAny := range patchArray {
		patchNode, ok := patchAny.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(patches, fmt.Sprintf("invalid patch item type: %T", patchAny))
		}

		patchID, err := GetManifestID(patchNode)
		if err != nil {
			return nil, NewEvaluationErrorWithParent(patchNode, fmt.Sprintf("failed to get GVK from patch"), err)
		}

		for i, targetAny := range targetArray {
			targetNode, ok := targetAny.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(targets, fmt.Sprintf("invalid target item type: %T", targetAny))
			}

			targetID, err := GetManifestID(targetNode)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(targetNode, fmt.Sprintf("failed to get GVK from target"), err)
			}

			if patchID == targetID {
				targetArray[i], err = DeepMergeYispNode(targetNode, patchNode, targetNode.Type)
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
		if node.Kind == KindArray {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid argument type for flatten: %T", node))
			}
			for _, item := range arr {
				itemNode, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				flattened = append(flattened, itemNode)
			}
		} else {
			flattened = append(flattened, node)
		}
	}

	return &YispNode{
		Kind:           KindArray,
		Value:          flattened,
		IsDocumentRoot: true,
	}, nil
}

func opAssertType(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("assert-type requires 2 arguments, got %d", len(cdr)))
	}

	schemaNode := cdr[0]
	valueNode := cdr[1]

	if schemaNode.Kind != KindType {
		return nil, NewEvaluationError(schemaNode, fmt.Sprintf("assert-type requires a type as the first argument, got %v", schemaNode.Kind))
	}
	schema, ok := schemaNode.Value.(*Schema)
	if !ok {
		return nil, NewEvaluationError(schemaNode, fmt.Sprintf("invalid type value: %T", schemaNode.Value))
	}

	err := schema.Validate(valueNode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(valueNode, fmt.Sprintf("value does not match schema: %s", err.Error()), err)
	}

	return valueNode, nil
}

func opGetType(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("get-type requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]

	if node.Type == nil {
		return &YispNode{
			Kind:  KindNull,
			Value: nil,
			Attr:  node.Attr,
		}, nil
	}

	return node.Type.ToYispNode()
}

func opTypeOf(cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error) {
	if len(cdr) != 1 {
		return nil, NewEvaluationError(nil, fmt.Sprintf("typeof requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]

	if node.Type == nil {
		return &YispNode{
			Kind:  KindNull,
			Value: nil,
			Attr:  node.Attr,
		}, nil
	}

	return &YispNode{
		Kind:  KindType,
		Value: node.Type,
		Attr:  node.Attr,
	}, nil
}
