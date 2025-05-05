package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/totegamma/yisp/yaml"
	"io"
	"maps"
	"strconv"
	"strings"
)

func JsonPrint(tag string, obj any) {
	b, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Println(tag, string(b))
}

type Environment struct {
	Vars map[string]any
}

func NewEnvironment() *Environment {
	return &Environment{
		Vars: make(map[string]any),
	}
}

type Kind int32

const (
	KindSymbol Kind = iota
	KindParameter
	KindNull
	KindBool
	KindInt
	KindFloat
	KindString
	KindArray
	KindMap
)

type YispNode struct {
	Kind   Kind
	Tag    string
	Body   any
	Params []string
}

func (env *Environment) Clone() *Environment {
	newEnv := NewEnvironment()
	maps.Copy(newEnv.Vars, env.Vars)
	return newEnv
}

func eval(node *YispNode, env *Environment) (any, error) {

	if node.Tag == "!discard" {
		return nil, nil
	}

	switch node.Kind {
	case KindSymbol:
		var ok bool
		var body any
		body, ok = globals.Vars[node.Body.(string)]
		if !ok {
			body, ok = env.Vars[node.Body.(string)]
			if !ok {
				return nil, fmt.Errorf("undefined symbol: %s", node.Body)
			}
		}
		node, ok := body.(*YispNode)
		if !ok {
			return body, nil
		}

		return eval(node, env)

	case KindParameter:
		return fmt.Sprintf("parameter: %v (type: %v)", node.Body, node.Tag), nil

	case KindNull:
		return nil, nil

	case KindBool:
		if node.Body == "true" {
			return true, nil
		}
		return false, nil

	case KindFloat:
		if f, err := strconv.ParseFloat(node.Body.(string), 64); err == nil {
			return f, nil
		}
		return nil, fmt.Errorf("invalid float value: %s", node.Body)

	case KindInt:
		if i, err := strconv.Atoi(node.Body.(string)); err == nil {
			return i, nil
		}
		return nil, fmt.Errorf("invalid int value: %s", node.Body)

	case KindString:
		return node.Body, nil

	case KindArray:

		arr, ok := node.Body.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array type: %T", node.Body)
		}

		if node.Tag == "!eval" {
			carNode, ok := arr[0].(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid car type: %T", arr[0])
			}
			car, ok := carNode.Body.(string)
			if !ok {
				return nil, fmt.Errorf("invalid car value: %T", carNode.Body)
			}

			cdr := arr[1:]

			switch car {
			case "join":
				var result string
				for _, item := range cdr {
					node, ok := item.(*YispNode)
					if !ok {
						return nil, fmt.Errorf("invalid item type: %T", item)
					}
					val, err := eval(node, env)
					if err != nil {
						return nil, err
					}
					str, ok := val.(string)
					if !ok {
						return nil, fmt.Errorf("invalid argument type for join: %T", val)
					}
					result += str
				}
				return result, nil
			case "lambda":
				if len(cdr) != 2 {
					return nil, fmt.Errorf("lambda requires 2 arguments")
				}

				paramsNode, ok := cdr[0].(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid params type: %T", cdr[0])
				}

				params := make([]string, 0)
				for _, item := range paramsNode.Body.([]any) {
					paramNode, ok := item.(*YispNode)
					if !ok {
						return nil, fmt.Errorf("invalid param type: %T", item)
					}
					param, ok := paramNode.Body.(string)
					if !ok {
						return nil, fmt.Errorf("invalid param value: %T", paramNode.Body)
					}
					params = append(params, param)
				}

				body, ok := cdr[1].(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid body type: %T", cdr[1])
				}

				body.Params = params

				return body, nil

			default:
				val, ok := globals.Vars[car]
				if !ok {
					return nil, fmt.Errorf("unknown function: %s", car)
				}

				function, ok := val.(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid function type: %T", val)
				}

				function.Tag = "!eval"

				funVal, err := eval(function, env)
				if err != nil {
					return nil, err
				}

				funNode, ok := funVal.(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid function value type: %T", funVal)
				}

				if len(funNode.Params) != len(cdr) {
					return nil, fmt.Errorf("function %s requires %d arguments, got %d", car, len(funNode.Params), len(cdr))
				}

				newEnv := env.Clone()

				for i, item := range cdr {
					node, ok := item.(*YispNode)
					if !ok {
						return nil, fmt.Errorf("invalid item type: %T", item)
					}
					val, err := eval(node, env)
					if err != nil {
						return nil, err
					}
					newEnv.Vars[funNode.Params[i]] = val
				}

				return eval(funNode, newEnv)
			}

		} else {

			results := make([]any, len(arr))
			for i, item := range arr {
				node, ok := item.(*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid item type: %T", item)
				}

				result, err := eval(node, env)
				if err != nil {
					return nil, err
				}
				results[i] = result
			}
			return results, nil
		}

	case KindMap:
		m, ok := node.Body.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid map type: %T", node.Body)
		}
		results := make(map[string]any)
		for key, item := range m {

			node, ok := item.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}

			val, err := eval(node, env)
			if err != nil {
				return nil, err
			}
			results[key] = val
		}
		return results, nil
	}
	return nil, nil
}

var globals = NewEnvironment()

func parse(node *yaml.Node, env *Environment) (*YispNode, error) {

	var result *YispNode
	var err error

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		result, err = parse(node.Content[0], env)

	case yaml.SequenceNode:
		s := make([]any, len(node.Content))
		for i, item := range node.Content {
			value, err := parse(item, env)
			if err != nil {
				return nil, err
			}
			s[i] = value
		}

		result = &YispNode{
			Kind: KindArray,
			Body: s,
			Tag:  node.Tag,
		}

	case yaml.MappingNode:
		m := make(map[string]any)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			key := keyNode.Value
			value, err := parse(valueNode, env)
			if err != nil {
				return nil, err
			}
			m[key] = value
		}

		result = &YispNode{
			Kind: KindMap,
			Body: m,
			Tag:  node.Tag,
		}

	case yaml.ScalarNode:

		var kind Kind
		switch node.Tag {
		case "!!null":
			kind = KindNull
		case "!!bool":
			kind = KindBool
		case "!!int":
			kind = KindInt
		case "!!float":
			kind = KindFloat
		case "!!str":
			kind = KindString
		case "!string", "!number", "!bool":
			kind = KindParameter
		}

		result = &YispNode{
			Kind: kind,
			Body: node.Value,
			Tag:  node.Tag,
		}

	case yaml.AliasNode:

		result = &YispNode{
			Kind: KindSymbol,
			Body: node.Value,
			Tag:  node.Tag,
		}
	}

	if node.Anchor != "" {
		globals.Vars[node.Anchor] = result
	}

	return result, err
}

func main() {

	data := `
!discard
&mkpod
- lambda
- - !string name
  - !string image
- apiVersion: v1
  kind: Pod
  metadata:
    name: *name
  spec:
    containers:
      - name: *name
        image: *image
---
!eval
- mkpod
- mypod1
- myimage1
`

	decoder := yaml.NewDecoder(strings.NewReader(data))
	if decoder == nil {
		panic("failed to create decoder")
	}

	for {
		var root yaml.Node
		err := decoder.Decode(&root)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		env := NewEnvironment()
		parsed, err := parse(&root, env)
		if err != nil {
			panic(err)
		}

		result, err := eval(parsed, globals)
		if err != nil {
			panic(err)
		}

		JsonPrint("result", result)

	}
}
