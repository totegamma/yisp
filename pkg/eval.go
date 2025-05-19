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
			val, err := Eval(node, env, mode)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate argument"), err)
			}

			if lambda.Arguments[i].Schema != nil {
				err := lambda.Arguments[i].Schema.Validate(val)
				if err != nil {
					return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("object does not satisfy type"), err)
				}
			}

			newEnv.Vars[lambda.Arguments[i].Name] = val
		}

		return Eval(lambda.Body, newEnv, mode)

	case KindString:
		return Call(car, cdr, env, mode)

	default:
		return nil, NewEvaluationError(car, fmt.Sprintf("cannot apply type %s", car.Kind))
	}
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

			carNode, ok := arr[0].(*YispNode)
			if !ok {
				return nil, NewEvaluationError(node, fmt.Sprintf("invalid car type: %T", arr[0]))
			}

			car, err := Eval(carNode, env, mode)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to evaluate car"), err)
			}

			cdr := make([]*YispNode, len(arr)-1)
			for i, item := range arr[1:] {
				node, ok := item.(*YispNode)
				if !ok {
					return nil, NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				cdr[i] = node
			}

			if showTrace {
				if car.Kind == KindLambda {
					fmt.Printf("%s->%s\n", pad(env.Depth()), car)
				} else {
					fmt.Printf("%s->%s\n", pad(env.Depth()), car.Value)
				}
			}

			r, err := Apply(car, cdr, env, mode)
			if err != nil {
				return nil, NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to apply function"), err)
			}

			if showTrace {
				if r.Kind == KindLambda {
					fmt.Printf("%s<-%s\n", pad(env.Depth()), r)
				} else {
					fmt.Printf("%s<-%s\n", pad(env.Depth()), r.Value)
				}
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

	if node.Anchor != "" {
		env.Set(node.Anchor, result)

		if result.Kind == KindLambda {
			lambda, ok := result.Value.(*Lambda)
			if ok {
				lambda.Clojure.Set(node.Anchor, result)
			}
		}
	}

	return result, nil
}
