package main

import (
	"strconv"
	"maps"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/totegamma/yisp/yaml"
	"io"
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
	Kind  Kind
	Tag   string
	Body  any
}


func (env *Environment) Clone() *Environment {
	newEnv := NewEnvironment()
	maps.Copy(newEnv.Vars, env.Vars)
	return newEnv
}

/*
func eval(name string, arg []any) (any, error) {
	switch name {
	case "join":
		var result string
		for _, item := range arg {
			if str, ok := item.(string); ok {
				result += str
			} else {
				return nil, fmt.Errorf("invalid argument type for join: %T", item)
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}
*/

func eval(node *YispNode, env *Environment) (any, error) {
	switch node.Kind {
	case KindSymbol:
		return fmt.Sprintf("symbol: %v", node.Body), nil

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
- null
- true
- 3
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

		//JsonPrint("parsed", parsed)

		result, err := eval(parsed, globals)
		if err != nil {
			panic(err)
		}

		JsonPrint("result", result)

		/*
		JsonPrint("parsed", parsed)
		JsonPrint("globals", globals)
		*/

	}
}
