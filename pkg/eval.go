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
		lambda, ok := car.Value.([]*YispNode)
		if !ok {
			return nil, fmt.Errorf("invalid lambda type: %T", car.Value)
		}

		if len(lambda) != 2 {
			return nil, fmt.Errorf("lambda requires 2 arguments")
		}

		paramsNode := lambda[0]
		bodyNode := lambda[1]

		params := make([]string, 0)
		for _, item := range paramsNode.Value.([]any) {
			paramNode, ok := item.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid param type: %T", item)
			}
			param, ok := paramNode.Value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid param value: %T", paramNode.Value)
			}
			params = append(params, param)
		}

		newEnv := env.CreateChild()
		for i, node := range cdr {
			val, err := Eval(node, env, mode)
			if err != nil {
				return nil, err
			}
			newEnv.Vars[params[i]] = val
		}

		return Eval(bodyNode, newEnv, mode)

	case KindString:
		op, ok := car.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid car value: %T", car.Value)
		}

		switch op {
		case "concat":
			var result string
			for _, node := range cdr {
				val, err := Eval(node, env, mode)
				if err != nil {
					return nil, err
				}
				str, ok := val.Value.(string)
				if !ok {
					return nil, fmt.Errorf("invalid argument type for concat: %T", val)
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
					return nil, err
				}
				num, ok := val.Value.(int)
				if !ok {
					return nil, fmt.Errorf("invalid argument type for +: %T", val)
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
				return nil, err
			}
			baseNum, ok := firstNode.Value.(int)
			if !ok {
				return nil, fmt.Errorf("invalid argument type for -: %T", firstNode)
			}
			for _, node := range cdr[1:] {
				evaluated, err := Eval(node, env, mode)
				if err != nil {
					return nil, err
				}
				val, ok := evaluated.Value.(int)
				if !ok {
					return nil, fmt.Errorf("invalid argument type for -: %T", evaluated)
				}
				baseNum -= val
			}
			return &YispNode{
				Kind:  KindInt,
				Value: baseNum,
			}, nil

		case "if":
			if len(cdr) != 3 {
				return nil, fmt.Errorf("if requires 3 arguments")
			}

			condNode, err := Eval(cdr[0], env, mode)
			if err != nil {
				return nil, err
			}

			cond, ok := condNode.Value.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid condition type: %T", condNode.Value)
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
					return nil, fmt.Errorf("invalid path type: %T", node.Value)
				}

				baseDir := filepath.Dir(node.File)
				joinedPath := filepath.Join(baseDir, relpath)
				path := filepath.Clean(joinedPath)

				var err error
				results[i], err = evaluateYisp(path, env.CreateChild())
				if err != nil {
					return nil, err
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
					return nil, fmt.Errorf("invalid tuple type: %T", node.Value)
				}

				if len(tuple) != 2 {
					return nil, fmt.Errorf("import requires 2 arguments")
				}

				nameNode, ok := tuple[0].(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid name type: %T", tuple[0])
				}

				name, ok := nameNode.Value.(string)
				if !ok {
					return nil, fmt.Errorf("invalid name type: %T", nameNode.Value)
				}

				relpathNode, ok := tuple[1].(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid path type: %T", tuple[1])
				}

				relpath, ok := relpathNode.Value.(string)
				if !ok {
					return nil, fmt.Errorf("invalid path type: %T", relpathNode.Value)
				}

				joinedPath := filepath.Join(baseDir, relpath)
				path := filepath.Clean(joinedPath)

				newEnv := NewEnv()

				var err error
				_, err = evaluateYisp(path, newEnv)
				if err != nil {
					return nil, err
				}

				env.AddModule(name, newEnv)

			}

			return &YispNode{
				Kind: KindNull,
			}, nil

		case "lambda":
			if len(cdr) != 2 {
				return nil, fmt.Errorf("lambda requires 2 arguments")
			}

			return &YispNode{
				Kind:  KindLambda,
				Value: cdr,
			}, nil

		default:
			return nil, fmt.Errorf("unknown function name: %s", op)
		}
	default:
		return nil, fmt.Errorf("invalid car type2: %T", car)
	}
}

func compareInts(cdr []*YispNode, env *Env, mode EvalMode, opName string, cmp func(int, int) bool) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, fmt.Errorf("%s requires 2 arguments", opName)
	}
	firstNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, err
	}
	firstNum, ok := firstNode.Value.(int)
	if !ok {
		// Attempt to convert float to int if applicable, or handle other types
		if firstFloat, isFloat := firstNode.Value.(float64); isFloat {
			firstNum = int(firstFloat) // Note: This truncates. Decide if this is the desired behavior.
		} else {
			return nil, fmt.Errorf("invalid first argument type for %s: %T (value: %v)", opName, firstNode.Value, firstNode.Value)
		}
	}

	secondNode, err := Eval(cdr[1], env, mode)
	if err != nil {
		return nil, err
	}
	secondNum, ok := secondNode.Value.(int)
	if !ok {
		if secondFloat, isFloat := secondNode.Value.(float64); isFloat {
			secondNum = int(secondFloat) // Note: This truncates.
		} else {
			return nil, fmt.Errorf("invalid second argument type for %s: %T (value: %v)", opName, secondNode.Value, secondNode.Value)
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

	switch node.Kind {
	case KindSymbol:
		var ok bool
		var body any

		body, ok = env.Get(node.Value.(string))
		if !ok {
			return nil, fmt.Errorf("undefined symbol: %s", node.Value)
		}
		node, ok := body.(*YispNode)
		if !ok {
			return nil, fmt.Errorf("invalid symbol type: %T", body)
		}

		return Eval(node, env, mode)

	case KindParameter:
		return node, nil

	case KindNull:
		return &YispNode{
			Kind:  KindNull,
			Value: nil,
			Tag:   node.Tag,
		}, nil

	case KindBool:
		val := false
		if node.Value == "true" {
			val = true
		}
		return &YispNode{
			Kind:  KindBool,
			Value: val,
			Tag:   node.Tag,
		}, nil

	case KindFloat:
		if f, err := strconv.ParseFloat(node.Value.(string), 64); err == nil {
			return &YispNode{
				Kind:  KindFloat,
				Value: f,
				Tag:   node.Tag,
			}, nil
		}
		return nil, fmt.Errorf("invalid float value: %s", node.Value)

	case KindInt:
		var i int
		var ok bool
		i, ok = node.Value.(int)
		if !ok {
			iStr, ok := node.Value.(string)
			if !ok {
				return nil, fmt.Errorf("invalid int type: %T", node.Value)
			}

			var err error
			i, err = strconv.Atoi(iStr)
			if err != nil {
				return nil, fmt.Errorf("invalid int value: %s", iStr)
			}
		}

		return &YispNode{
			Kind:  KindInt,
			Value: i,
			Tag:   node.Tag,
		}, nil

	case KindString:
		return &YispNode{
			Kind:  KindString,
			Value: node.Value,
			Tag:   node.Tag,
		}, nil

	case KindArray:
		if mode == EvalModeEval {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, fmt.Errorf("invalid array type: %T", node.Value)
			}

			carNode, ok := arr[0].(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid car type1: %T", arr[0])
			}

			car, err := Eval(carNode, env, mode)
			if err != nil {
				return nil, err
			}

			cdr := make([]*YispNode, len(arr)-1)
			for i, item := range arr[1:] {
				node, ok := item.(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid item type: %T", item)
				}
				cdr[i] = node
			}

			return Apply(car, cdr, env, mode)
		} else {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, fmt.Errorf("invalid array type: %T", node.Value)
			}

			results := make([]any, len(arr))
			for i, item := range arr {
				node, ok := item.(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid item type: %T", item)
				}

				result, err := Eval(node, env, mode)
				if err != nil {
					return nil, err
				}
				results[i] = result
			}
			return &YispNode{
				Kind:  KindArray,
				Value: results,
				Tag:   node.Tag,
			}, nil
		}

	case KindMap:
		m, ok := node.Value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid map type: %T", node.Value)
		}
		results := make(map[string]any)
		for key, item := range m {
			node, ok := item.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}

			val, err := Eval(node, env, mode)
			if err != nil {
				return nil, err
			}
			results[key] = val
		}

		return &YispNode{
			Kind:  KindMap,
			Value: results,
			Tag:   node.Tag,
		}, nil
	}
	return nil, nil
}
