package yisp

import (
	"fmt"
	"path/filepath"
	"strconv"
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
			val, err := Eval(node, env, mode)
			if err != nil {
				return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate argument: %s", err))
			}
			newEnv.Vars[lambda.Params[i]] = val
		}

		return Eval(lambda.Body, newEnv, mode)

	case KindString:
		op, ok := car.Value.(string)
		if !ok {
			return nil, NewEvaluationError(car, fmt.Sprintf("invalid car value: %T", car.Value))
		}

		switch op {
		case "concat":
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

		case "+":
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

		case "-":
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

		case "*":
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
		case "/":
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

		case "if":
			if len(cdr) != 3 {
				return nil, NewEvaluationError(car, fmt.Sprintf("if requires 3 arguments, got %d", len(cdr)))
			}

			condNode, err := Eval(cdr[0], env, mode)
			if err != nil {
				return nil, NewEvaluationError(cdr[0], fmt.Sprintf("failed to evaluate condition: %s", err))
			}

			cond := false
			switch condNode.Value.(type) {
			case bool:
				cond, ok = condNode.Value.(bool)
			case int:
				condInt, ok := condNode.Value.(int)
				if !ok {
					return nil, NewEvaluationError(condNode, fmt.Sprintf("invalid condition type: %T", condNode.Value))
				}
				cond = condInt != 0
			case float64:
				condFloat, ok := condNode.Value.(float64)
				if !ok {
					return nil, NewEvaluationError(condNode, fmt.Sprintf("invalid condition type: %T", condNode.Value))
				}
				cond = condFloat != 0.0
			case string:
				condStr, ok := condNode.Value.(string)
				if !ok {
					return nil, NewEvaluationError(condNode, fmt.Sprintf("invalid condition type: %T", condNode.Value))
				}
				cond = condStr != ""
			case []any:
				condArr, ok := condNode.Value.([]any)
				if !ok {
					return nil, NewEvaluationError(condNode, fmt.Sprintf("invalid condition type: %T", condNode.Value))
				}
				cond = len(condArr) != 0
			case map[string]any:
				condMap, ok := condNode.Value.(map[string]any)
				if !ok {
					return nil, NewEvaluationError(condNode, fmt.Sprintf("invalid condition type: %T", condNode.Value))
				}
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

		case "==":
			return compareInts(cdr, env, mode, "==", func(a, b int) bool { return a == b })
		case "!=":
			return compareInts(cdr, env, mode, "!=", func(a, b int) bool { return a != b })
		case "<":
			return compareInts(cdr, env, mode, "<", func(a, b int) bool { return a < b })
		case "<=":
			return compareInts(cdr, env, mode, "<=", func(a, b int) bool { return a <= b })
		case ">":
			return compareInts(cdr, env, mode, ">", func(a, b int) bool { return a > b })
		case ">=":
			return compareInts(cdr, env, mode, ">=", func(a, b int) bool { return a >= b })

		case "car":
			if len(cdr) != 1 {
				return nil, NewEvaluationError(car, fmt.Sprintf("car requires 1 argument, got %d", len(cdr)))
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

		case "cdr":
			if len(cdr) != 1 {
				return nil, NewEvaluationError(car, fmt.Sprintf("cdr requires 1 argument, got %d", len(cdr)))
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

		case "cons":
			if len(cdr) != 2 {
				return nil, NewEvaluationError(car, fmt.Sprintf("cons requires 2 arguments, got %d", len(cdr)))
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

		case "discard":
			for _, node := range cdr {
				Eval(node, env, mode)
			}

			return nil, nil

		case "include":
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

		case "import":
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

		case "lambda":
			if len(cdr) != 2 {
				return nil, NewEvaluationError(car, fmt.Sprintf("lambda requires 2 arguments, got %d", len(cdr)))
			}

			paramsNode := cdr[0]
			bodyNode := cdr[1]

			params := make([]string, 0)
			for _, item := range paramsNode.Value.([]any) {
				paramNode, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(car, fmt.Sprintf("invalid param type: %T", item))
				}
				param, ok := paramNode.Value.(string)
				if !ok {
					return nil, NewEvaluationError(car, fmt.Sprintf("invalid param value: %T", paramNode.Value))
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

		default:
			JsonPrint("env", env)
			return nil, NewEvaluationError(car, fmt.Sprintf("unknown function name: %s", op))
		}
	default:
		return nil, NewEvaluationError(car, fmt.Sprintf("cannot apply type %s", car.Kind))
	}
}

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

// Eval evaluates a YispNode in the given environment
func Eval(node *YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	if node.Tag == "!yisp" {
		mode = EvalModeEval
	}

	if node.Tag == "!quote" {
		mode = EvalModeQuote
	}

	var result *YispNode

	switch node.Kind {
	case KindSymbol:
		var ok bool
		var body any

		body, ok = env.Get(node.Value.(string))
		if !ok {
			JsonPrint("env", env)
			return nil, NewEvaluationError(node, fmt.Sprintf("undefined symbol: %s", node.Value))
		}
		node, ok := body.(*YispNode)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid symbol type: %T", body))
		}

		result = node

	case KindParameter:
		result = node

	case KindNull:
		result = &YispNode{
			Kind:  KindNull,
			Value: nil,
			Tag:   node.Tag,
		}

	case KindBool:
		val := false
		if node.Value == "true" {
			val = true
		}
		result = &YispNode{
			Kind:  KindBool,
			Value: val,
			Tag:   node.Tag,
		}

	case KindFloat:
		var f float64
		var ok bool
		f, ok = node.Value.(float64)
		if !ok {
			fStr, ok := node.Value.(string)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid float type: %T", node.Value))
			}

			var err error
			f, err = strconv.ParseFloat(fStr, 64)
			if err != nil {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid float value: %s", fStr))
			}
		}

		result = &YispNode{
			Kind:  KindFloat,
			Value: f,
			Tag:   node.Tag,
		}

	case KindInt:
		var i int
		var ok bool
		i, ok = node.Value.(int)
		if !ok {
			iStr, ok := node.Value.(string)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid int type: %T", node.Value))
			}

			var err error
			i, err = strconv.Atoi(iStr)
			if err != nil {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid int value: %s", iStr))
			}
		}

		result = &YispNode{
			Kind:  KindInt,
			Value: i,
			Tag:   node.Tag,
		}

	case KindString:
		result = &YispNode{
			Kind:  KindString,
			Value: node.Value,
			Tag:   node.Tag,
		}

	case KindArray:
		if mode == EvalModeEval {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid array type: %T", node.Value))
			}

			carNode, ok := arr[0].(*YispNode)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid car type: %T", arr[0]))
			}

			car, err := Eval(carNode, env, mode)
			if err != nil {
				return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate car: %s", err))
			}

			cdr := make([]*YispNode, len(arr)-1)
			for i, item := range arr[1:] {
				node, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				cdr[i] = node
			}

			r, err := Apply(car, cdr, env, mode)
			if err != nil {
				return nil, NewEvaluationError(node, fmt.Sprintf("failed to apply function: %s", err))
			}
			result = r

		} else {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid array type: %T", node.Value))
			}

			results := make([]any, len(arr))
			for i, item := range arr {
				node, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}

				result, err := Eval(node, env, mode)
				if err != nil {
					return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate item: %s", err))
				}
				results[i] = result
			}
			result = &YispNode{
				Kind:  KindArray,
				Value: results,
				Tag:   node.Tag,
			}
		}

	case KindMap:
		m, ok := node.Value.(map[string]any)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid map type: %T", node.Value))
		}
		results := make(map[string]any)
		for key, item := range m {
			node, ok := item.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
			}

			val, err := Eval(node, env, mode)
			if err != nil {
				return nil, NewEvaluationError(node, fmt.Sprintf("failed to evaluate item: %s", err))
			}
			results[key] = val
		}

		result = &YispNode{
			Kind:  KindMap,
			Value: results,
			Tag:   node.Tag,
		}
	}

	if node.Anchor != "" {
		env.Set(node.Anchor, result)
	}

	return result, nil
}
