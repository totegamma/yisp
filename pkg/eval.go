package yisp

import (
	"fmt"
	"strconv"
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
		return Call(car, cdr, env.CreateChild(), mode)

	default:
		return nil, NewEvaluationError(car, fmt.Sprintf("cannot apply type %s", car.Kind))
	}
}

// Eval evaluates a YispNode in the given environment
func Eval(node *YispNode, env *Env, mode EvalMode) (*YispNode, error) {

	if showTrace {
		val, err := ToNative(node)
		if err != nil {
			return nil, NewEvaluationError(node, fmt.Sprintf("failed to convert node to native: %v", err))
		}
		fmt.Printf("%sEVAL: %v\n", pad(env.Depth()), val)
	}

	if node.Tag == "!yisp" {
		mode = EvalModeEval
	}

	if node.Tag == "!quote" {
		mode = EvalModeQuote
	}

	result := node

	switch node.Kind {
	case KindSymbol:
		var ok bool
		var body any

		body, ok = env.Get(node.Value.(string))
		if !ok {
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
			Pos:   node.Pos,
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
			Pos:   node.Pos,
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
			Pos:   node.Pos,
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
			Pos:   node.Pos,
		}

	case KindString:
		result = &YispNode{
			Kind:  KindString,
			Value: node.Value,
			Tag:   node.Tag,
			Pos:   node.Pos,
		}

	case KindArray:
		if mode == EvalModeEval {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid array type: %T", node.Value))
			}

			if len(arr) == 0 {
				goto END_EVAL
			}

			nodes := make([]*YispNode, len(arr))
			for i, item := range arr {
				node, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				nodes[i] = node
			}

			// check special forms
			op, ok := nodes[0].Value.(string)
			if ok {
				switch op {
				case "if":
					if len(nodes) != 4 {
						return nil, NewEvaluationError(nodes[0], "if requires 3 arguments")
					}
					condNode, err := Eval(nodes[1], env, mode)
					if err != nil {
						return nil, NewEvaluationErrorWithParent(nodes[1], "failed to evaluate condition", err)
					}

					cond, err := isTruthy(condNode)
					if err != nil {
						return nil, NewEvaluationErrorWithParent(nodes[1], "failed to evaluate condition", err)
					}

					if cond {
						result, err = Eval(nodes[2], env, mode)
						if err != nil {
							return nil, NewEvaluationErrorWithParent(nodes[2], "failed to evaluate true branch", err)
						}
						goto END_EVAL
					} else {
						result, err = Eval(nodes[3], env, mode)
						if err != nil {
							return nil, NewEvaluationErrorWithParent(nodes[3], "failed to evaluate false branch", err)
						}
						goto END_EVAL
					}
				case "lambda":
					if len(nodes) < 3 {
						return nil, NewEvaluationError(nodes[0], "lambda requires at least 2 arguments")
					}

					paramsNode := nodes[1]
					bodyNode := nodes[2]

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

					result = &YispNode{
						Kind:  KindLambda,
						Value: lambda,
						Tag:   node.Tag,
						Pos:   node.Pos,
					}
					goto END_EVAL
				case "import":
					for _, node := range nodes[1:] {

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

					result = &YispNode{
						Kind: KindNull,
					}
					goto END_EVAL
				}
			}

			evaluated := make([]*YispNode, len(nodes))
			for i, item := range nodes {
				e, err := Eval(item, env, mode)
				if err != nil {
					return nil, NewEvaluationErrorWithParent(item, fmt.Sprintf("failed to evaluate item %d", i), err)
				}
				evaluated[i] = e
			}

			var err error
			result, err = Apply(evaluated[0], evaluated[1:], env, mode)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to apply function"), err)
			}

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
					return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate item"), err)
				}
				results[i] = result
			}
			result = &YispNode{
				Kind:  KindArray,
				Value: results,
				Tag:   node.Tag,
				Pos:   node.Pos,
			}
		}

	case KindMap:
		m, ok := node.Value.(*YispMap)
		if !ok {
			return nil, NewEvaluationError(node, fmt.Sprintf("invalid map type: %T", node.Value))
		}
		results := NewYispMap()
		for key, item := range m.AllFromFront() {
			node, ok := item.(*YispNode)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
			}

			val, err := Eval(node, env, mode)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate item"), err)
			}

			if strings.HasPrefix(key, YISP_SPECIAL_MERGE_KEY) {
				if val.Kind == KindMap {
					innerMap, ok := val.Value.(*YispMap)
					if !ok {
						return nil, NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", val.Value))
					}
					for k, v := range innerMap.AllFromFront() {
						results.Set(k, v)
					}
				} else if val.Kind == KindArray {
					for _, item := range val.Value.([]any) {
						node, ok := item.(*YispNode)
						if !ok {
							return nil, NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", item))
						}
						if node.Kind == KindMap {
							innerMap, ok := node.Value.(*YispMap)
							if !ok {
								return nil, NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", node.Value))
							}
							for innerK, innerV := range innerMap.AllFromFront() {
								results.Set(innerK, innerV)
							}
						} else if node.Kind == KindNull {
							continue
						} else {
							return nil, NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", item))
						}
					}
				} else if val.Kind == KindNull {
					continue
				} else {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", val))
				}
			} else {
				results.Set(key, val)
			}

		}

		result = &YispNode{
			Kind:  KindMap,
			Value: results,
			Tag:   node.Tag,
			Pos:   node.Pos,
		}
	}

END_EVAL:

	if node.Anchor != "" {
		env.Root().Set(node.Anchor, result) // anchor is global

		if result.Kind == KindLambda {
			lambda, ok := result.Value.(*Lambda)
			if ok {
				lambda.Clojure.Set(node.Anchor, result)
			}
		}
	}

	if node.Tag != "" && node.Tag != "!quote" && node.Tag != "!yisp" {
		typeName := strings.TrimPrefix(node.Tag, "!")
		if typeName != "" && !strings.HasPrefix(typeName, "!") {
			typeNode, ok := env.Get(typeName)
			if ok && typeNode.Kind == KindType {
				schema, ok := typeNode.Value.(*Schema)
				if ok {
					casted, err := schema.Cast(result)
					if err != nil {
						return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to cast to type %s", typeName), err)
					}
					result = casted
				}
			}
		}
	}

	return result, nil
}
