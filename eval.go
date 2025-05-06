package yisp

import (
	"fmt"
	"strconv"
)

// Apply applies a function to arguments
func Apply(car *YispNode, cdr []*YispNode, env *Env) (*YispNode, error) {
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
			val, err := Eval(node, env)
			if err != nil {
				return nil, err
			}
			newEnv.Vars[params[i]] = val
		}

		return Eval(bodyNode, newEnv)

	case KindString:
		op, ok := car.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid car value: %T", car.Value)
		}

		switch op {
		case "join":
			var result string
			for _, node := range cdr {
				val, err := Eval(node, env)
				if err != nil {
					return nil, err
				}
				str, ok := val.Value.(string)
				if !ok {
					return nil, fmt.Errorf("invalid argument type for join: %T", val)
				}
				result += str
			}

			return &YispNode{
				Kind:  KindString,
				Value: result,
			}, nil

		case "discard":
			for _, node := range cdr {
				Eval(node, env)
			}

			return nil, nil

		case "include":
			results := make([]*YispNode, len(cdr))
			for i, node := range cdr {
				path, ok := node.Value.(string)
				if !ok {
					return nil, fmt.Errorf("invalid path type: %T", node.Value)
				}

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
		return nil, fmt.Errorf("invalid car type: %T", car)
	}
}

// Eval evaluates a YispNode in the given environment
func Eval(node *YispNode, env *Env) (*YispNode, error) {
	switch node.Kind {
	case KindSymbol:
		var ok bool
		var body any

		body, ok = env.Get(node.Value.(string))
		node, ok := body.(*YispNode)
		if !ok {
			return nil, fmt.Errorf("invalid symbol type: %T", body)
		}

		return Eval(node, env)

	case KindParameter:
		return node, nil

	case KindNull:
		return nil, nil

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
		if i, err := strconv.Atoi(node.Value.(string)); err == nil {
			return &YispNode{
				Kind:  KindInt,
				Value: i,
				Tag:   node.Tag,
			}, nil
		}
		return nil, fmt.Errorf("invalid int value: %s", node.Value)

	case KindString:
		return &YispNode{
			Kind:  KindString,
			Value: node.Value,
			Tag:   node.Tag,
		}, nil

	case KindArray:
		if node.Tag == "!yisp" {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, fmt.Errorf("invalid array type: %T", node.Value)
			}

			carNode, ok := arr[0].(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid car type: %T", arr[0])
			}

			car, err := Eval(carNode, env)
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

			return Apply(car, cdr, env)
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

				result, err := Eval(node, env)
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

			val, err := Eval(node, env)
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
