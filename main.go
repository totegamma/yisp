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

func YamlPrint(obj any) {
	b, _ := yaml.Marshal(obj)
	fmt.Println(string(b))
}

type Environment struct {
	Vars map[string]*YispNode
}

func NewEnvironment() *Environment {
	return &Environment{
		Vars: make(map[string]*YispNode),
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
	KindLambda
)

type YispNode struct {
	Kind  Kind
	Tag   string
	Value any
}

func (env *Environment) Clone() *Environment {
	newEnv := NewEnvironment()
	maps.Copy(newEnv.Vars, env.Vars)
	return newEnv
}

func apply(car *YispNode, cdr []*YispNode, env *Environment) (*YispNode, error) {

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

		newEnv := env.Clone()
		for i, node := range cdr {
			val, err := eval(node, env)
			if err != nil {
				return nil, err
			}
			newEnv.Vars[params[i]] = val
		}

		return eval(bodyNode, newEnv)

	case KindString:
		op, ok := car.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid car value: %T", car.Value)
		}

		switch op {
		case "join":
			var result string
			for _, node := range cdr {
				val, err := eval(node, env)
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
				eval(node, env)
			}

			return nil, nil

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

func eval(node *YispNode, env *Environment) (*YispNode, error) {

	switch node.Kind {
	case KindSymbol:
		var ok bool
		var body any
		body, ok = globals.Vars[node.Value.(string)]
		if !ok {
			body, ok = env.Vars[node.Value.(string)]
			if !ok {
				return nil, fmt.Errorf("undefined symbol: %s", node.Value)
			}
		}
		node, ok := body.(*YispNode)
		if !ok {
			return nil, fmt.Errorf("invalid symbol type: %T", body)
		}

		return eval(node, env)

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

			car, err := eval(carNode, env)
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

			return apply(car, cdr, env)

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

				result, err := eval(node, env)
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

			val, err := eval(node, env)
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
			Kind:  KindArray,
			Value: s,
			Tag:   node.Tag,
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
			Kind:  KindMap,
			Value: m,
			Tag:   node.Tag,
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
			Kind:  kind,
			Value: node.Value,
			Tag:   node.Tag,
		}

	case yaml.AliasNode:

		result = &YispNode{
			Kind:  KindSymbol,
			Value: node.Value,
			Tag:   node.Tag,
		}
	}

	if node.Anchor != "" {
		globals.Vars[node.Anchor] = result
	}

	return result, err
}

func render(node *YispNode) any {
	switch node.Kind {
	case KindNull, KindBool, KindInt, KindFloat, KindString:
		return node.Value
	case KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return nil
		}
		results := make([]any, len(arr))
		for i, item := range arr {
			node, ok := item.(*YispNode)
			if !ok {
				return nil
			}
			results[i] = render(node)
		}
		return results
	case KindMap:
		m, ok := node.Value.(map[string]any)
		if !ok {
			return nil
		}
		results := make(map[string]any)
		for key, item := range m {
			node, ok := item.(*YispNode)
			if !ok {
				return nil
			}
			results[key] = render(node)
		}
		return results
	case KindLambda:
		return "(lambda)"
	case KindParameter:
		return "(parameter)"
	case KindSymbol:
		return "(symbol)"
	default:
		return "(unknown)"
	}
}

func main() {

	data := `
!yisp
&mkpod
- lambda
- [!string name, !string image]
- apiVersion: v1
  kind: Pod
  metadata:
    name: *name
  spec:
    containers:
      - name: *name
        image: *image
---
!yisp
- *mkpod
- mypod1
- myimage1
---
message: this is a normal yaml document
fruits:
  - apple
  - banana
  - chocolate
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

		evaluated, err := eval(parsed, globals)
		if err != nil {
			panic(err)
		}

		result := render(evaluated)

		YamlPrint(result)
		fmt.Println("---")
	}
}
