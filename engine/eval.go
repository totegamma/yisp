package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/totegamma/yisp/core"
)

// Eval evaluates a core.YispNode in the given environment
func (e *engine) Eval(node *core.YispNode, env *core.Env, mode core.EvalMode) (*core.YispNode, error) {

	if e.showTrace {
		val, err := node.ToNative()
		if err != nil {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("failed to convert node to native: %v", err))
		}
		fmt.Printf("%sEVAL: %v\n", pad(env.Depth()), val)
	}

	if node.Tag == "!yisp" {
		mode = core.EvalModeEval
	}

	if node.Tag == "!quote" {
		mode = core.EvalModeQuote
	}

	result := node

	switch node.Kind {
	case core.KindSymbol:
		var ok bool
		var body any

		body, ok = env.Get(node.Value.(string))
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("undefined symbol: %s", node.Value))
		}
		node, ok := body.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid symbol type: %T", body))
		}

		result = node

	case core.KindParameter:
		result = node

	case core.KindNull:
		result = &core.YispNode{
			Kind:  core.KindNull,
			Value: nil,
			Tag:   node.Tag,
			Attr:  node.Attr,
		}

	case core.KindBool:
		val := false
		if node.Value == "true" {
			val = true
		}
		result = &core.YispNode{
			Kind:  core.KindBool,
			Value: val,
			Tag:   node.Tag,
			Attr:  node.Attr,
		}

	case core.KindFloat:
		var f float64
		var ok bool
		f, ok = node.Value.(float64)
		if !ok {
			fStr, ok := node.Value.(string)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid float type: %T", node.Value))
			}

			var err error
			f, err = strconv.ParseFloat(fStr, 64)
			if err != nil {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid float value: %s", fStr))
			}
		}

		result = &core.YispNode{
			Kind:  core.KindFloat,
			Value: f,
			Tag:   node.Tag,
			Attr:  node.Attr,
		}

	case core.KindInt:
		var i int
		var ok bool
		i, ok = node.Value.(int)
		if !ok {
			iStr, ok := node.Value.(string)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid int type: %T", node.Value))
			}

			var err error
			i, err = strconv.Atoi(iStr)
			if err != nil {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid int value: %s", iStr))
			}
		}

		result = &core.YispNode{
			Kind:  core.KindInt,
			Value: i,
			Tag:   node.Tag,
			Attr:  node.Attr,
		}

	case core.KindString:
		result = &core.YispNode{
			Kind:  core.KindString,
			Value: node.Value,
			Tag:   node.Tag,
			Attr:  node.Attr,
		}

	case core.KindArray:
		if mode == core.EvalModeEval {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid array type: %T", node.Value))
			}

			if len(arr) == 0 {
				break
			}

			nodes := make([]*core.YispNode, len(arr))
			for i, item := range arr {
				node, ok := item.(*core.YispNode)
				if !ok {
					return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				nodes[i] = node
			}

			// check special forms
			op, ok := nodes[0].Value.(string)
			if !ok {
				op = ""
			}
			switch op {
			case "if":
				if len(nodes) != 4 {
					return nil, core.NewEvaluationError(nodes[0], "if requires 3 arguments")
				}
				condNode, err := e.Eval(nodes[1], env, mode)
				if err != nil {
					return nil, core.NewEvaluationErrorWithParent(nodes[1], "failed to evaluate condition", err)
				}

				cond, err := core.IsTruthy(condNode)
				if err != nil {
					return nil, core.NewEvaluationErrorWithParent(nodes[1], "failed to evaluate condition", err)
				}

				if cond {
					result, err = e.Eval(nodes[2], env, mode)
					if err != nil {
						return nil, core.NewEvaluationErrorWithParent(nodes[2], "failed to evaluate true branch", err)
					}
				} else {
					result, err = e.Eval(nodes[3], env, mode)
					if err != nil {
						return nil, core.NewEvaluationErrorWithParent(nodes[3], "failed to evaluate false branch", err)
					}
				}
			case "lambda":
				if len(nodes) < 3 {
					return nil, core.NewEvaluationError(nodes[0], "lambda requires at least 2 arguments")
				}

				paramsNode := nodes[1]
				bodyNode := nodes[2]

				params := make([]core.TypedSymbol, 0)
				for _, item := range paramsNode.Value.([]any) {
					paramNode, ok := item.(*core.YispNode)
					if !ok {
						return nil, core.NewEvaluationError(nil, fmt.Sprintf("invalid param type: %T", item))
					}
					param, ok := paramNode.Value.(string)
					if !ok {
						return nil, core.NewEvaluationError(nil, fmt.Sprintf("invalid param value: %T", paramNode.Value))
					}

					var schema *core.Schema
					tag := paramNode.Tag
					typeName := strings.TrimPrefix(tag, "!")
					if typeName != "" && !strings.HasPrefix(typeName, "!") {
						typeNode, ok := env.Get(typeName)
						if !ok {
							return nil, core.NewEvaluationError(nil, fmt.Sprintf("undefined type: %s", typeName))
						}
						if typeNode.Kind != core.KindType {
							return nil, core.NewEvaluationError(nil, fmt.Sprintf("%s is not a type. actual: %s", typeName, typeNode.Kind))
						}
						schema, ok = typeNode.Value.(*core.Schema)
						if !ok {
							return nil, core.NewEvaluationError(nil, fmt.Sprintf("invalid type value: %T", typeNode.Value))
						}
					}

					params = append(params, core.TypedSymbol{
						Name:   param,
						Schema: schema,
					})
				}

				var schema *core.Schema
				tag := paramsNode.Tag
				typeName := strings.TrimPrefix(tag, "!")
				if typeName != "" && !strings.HasPrefix(typeName, "!") {
					typeNode, ok := env.Get(typeName)
					if !ok {
						return nil, core.NewEvaluationError(nil, fmt.Sprintf("undefined type: %s", typeName))
					}
					if typeNode.Kind != core.KindType {
						return nil, core.NewEvaluationError(nil, fmt.Sprintf("%s is not a type. actual: %s", typeName, typeNode.Kind))
					}
					schema, ok = typeNode.Value.(*core.Schema)
					if !ok {
						return nil, core.NewEvaluationError(nil, fmt.Sprintf("invalid type value: %T", typeNode.Value))
					}
				}

				result = &core.YispNode{
					Kind: core.KindLambda,
					Value: &core.Lambda{
						Arguments: params,
						Returns:   schema,
						Body:      bodyNode,
						Clojure:   env.Clone(),
					},
					Tag:  node.Tag,
					Attr: node.Attr,
				}
			case "import":
				for _, node := range nodes[1:] {

					tuple, ok := node.Value.([]any)
					if !ok {
						return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid tuple type: %T", node.Value))
					}

					if len(tuple) != 2 {
						return nil, core.NewEvaluationError(node, fmt.Sprintf("import requires 2 arguments, got %d", len(tuple)))
					}

					nameNode, ok := tuple[0].(*core.YispNode)
					if !ok {
						return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid name type: %T", tuple[0]))
					}

					name, ok := nameNode.Value.(string)
					if !ok {
						return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid name type: %T", nameNode.Value))
					}

					relpathNode, ok := tuple[1].(*core.YispNode)
					if !ok {
						return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid path type: %T", tuple[1]))
					}

					relpath, ok := relpathNode.Value.(string)
					if !ok {
						return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid path type: %T", relpathNode.Value))
					}

					newEnv := core.NewEnv()

					var err error
					_, err = core.CallEngineByPath(relpath, node.Attr.File(), newEnv, e)
					if err != nil {
						return nil, core.NewEvaluationErrorWithParent(node, "failed to include file", err)
					}

					env.Root().Set(name, &core.YispNode{
						Kind:  core.KindMap,
						Value: newEnv.Vars,
					})
				}

				result = &core.YispNode{
					Kind: core.KindNull,
				}
			default:
				evaluated := make([]*core.YispNode, len(nodes))
				for i, item := range nodes {
					e, err := e.Eval(item, env, mode)
					if err != nil {
						return nil, core.NewEvaluationErrorWithParent(item, fmt.Sprintf("failed to evaluate item %d", i), err)
					}
					evaluated[i] = e
				}

				var err error
				result, err = e.Apply(evaluated[0], evaluated[1:], env, mode)
				if err != nil {
					return nil, core.NewEvaluationErrorWithParent(node, "failed to apply function", err)
				}
			}

		} else {
			arr, ok := node.Value.([]any)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid array type: %T", node.Value))
			}

			results := make([]any, len(arr))
			for i, item := range arr {
				node, ok := item.(*core.YispNode)
				if !ok {
					return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}

				result, err := e.Eval(node, env, mode)
				if err != nil {
					return nil, core.NewEvaluationErrorWithParent(node, "failed to evaluate item", err)
				}
				results[i] = result
			}
			result = &core.YispNode{
				Kind:  core.KindArray,
				Value: results,
				Tag:   node.Tag,
				Attr:  node.Attr,
			}
		}

	case core.KindMap:
		m, ok := node.Value.(*core.YispMap)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid map type: %T", node.Value))
		}
		results := core.NewYispMap()
		var schemaID string
		var apiVersion string
		var kind string
		for key, item := range m.AllFromFront() {
			node, ok := item.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
			}

			val, err := e.Eval(node, env, mode)
			if err != nil {
				return nil, core.NewEvaluationErrorWithParent(node, "failed to evaluate item", err)
			}

			switch key {
			case "$schema":
				schemaID, _ = val.Value.(string)
			case "apiVersion":
				apiVersion, _ = val.Value.(string)
			case "kind":
				kind, _ = val.Value.(string)
			}

			if strings.HasPrefix(key, core.YISP_SPECIAL_MERGE_KEY) {
				if val.Kind == core.KindMap {
					innerMap, ok := val.Value.(*core.YispMap)
					if !ok {
						return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", val.Value))
					}
					for k, v := range innerMap.AllFromFront() {
						results.Set(k, v)
					}
				} else if val.Kind == core.KindArray {
					for _, item := range val.Value.([]any) {
						node, ok := item.(*core.YispNode)
						if !ok {
							return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", item))
						}
						if node.Kind == core.KindMap {
							innerMap, ok := node.Value.(*core.YispMap)
							if !ok {
								return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", node.Value))
							}
							for innerK, innerV := range innerMap.AllFromFront() {
								results.Set(innerK, innerV)
							}
						} else if node.Kind == core.KindNull {
							continue
						} else {
							return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", item))
						}
					}
				} else if val.Kind == core.KindNull {
					continue
				} else {
					return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid merge item type: %T", val))
				}
			} else {
				results.Set(key, val)
			}
		}

		result = &core.YispNode{
			Kind:  core.KindMap,
			Value: results,
			Tag:   node.Tag,
			Attr:  node.Attr,
		}

		if schemaID != "" {
			schema, err := core.LoadSchemaFromURL(schemaID)
			if err != nil && !e.allowUntypedManifest {
				return nil, core.NewEvaluationError(
					node,
					fmt.Sprintf("failed to resolve type for %s.", schemaID),
				)
			}

			err = schema.Validate(result)
			if err != nil && !e.allowUntypedManifest {
				return nil, core.NewEvaluationErrorWithParent(
					node,
					fmt.Sprintf("manifest does not conform to schema %s: %s", schemaID, err.Error()),
					err,
				)
			}

			result.Type = schema
		} else if apiVersion != "" && kind != "" {
			split := strings.Split(apiVersion, "/")
			var group string
			var version string
			if len(split) == 1 {
				version = split[0]
			}
			if len(split) == 2 {
				group = split[0]
				version = split[1]
			}
			schema, err := core.LoadSchemaFromGVK(group, version, kind)
			if err != nil && !e.allowUntypedManifest {
				return nil, core.NewEvaluationError(
					node,
					fmt.Sprintf(
						"failed to resolve type for %s/%s/%s. Did you run `yisp cache-kube-schemas` first?",
						group,
						version,
						kind,
					),
				)
			}
			result.Type = schema
		}
	}

	if node.Anchor != "" {
		env.Root().Set(node.Anchor, result) // anchor is global

		if result.Kind == core.KindLambda {
			lambda, ok := result.Value.(*core.Lambda)
			if ok {
				lambda.Clojure.Set(node.Anchor, result)
			}
		}
	}

	if node.Tag != "" && node.Tag != "!quote" && node.Tag != "!yisp" {
		typeName := strings.TrimPrefix(node.Tag, "!")
		if typeName != "" && !strings.HasPrefix(typeName, "!") {
			typeNode, ok := env.Get(typeName)
			if ok && typeNode.Kind == core.KindType {
				schema, ok := typeNode.Value.(*core.Schema)
				if ok {
					casted, err := schema.Cast(result)
					if err != nil {
						return nil, core.NewEvaluationErrorWithParent(node, fmt.Sprintf("failed to cast to type %s", typeName), err)
					}
					result = casted
				}
			}
		}
	}

	return result, nil
}
