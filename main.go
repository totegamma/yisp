package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"github.com/pkg/errors"
	"github.com/totegamma/yisp/yaml"
)

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

func bake(node *yaml.Node) (any, error) {
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return bake(node.Content[0])
	case yaml.SequenceNode:
		s := make([]any, len(node.Content))
		for i, item := range node.Content {
			value, err := bake(item)
			if err != nil {
				return nil, err
			}
			s[i] = value
		}

		if node.Tag == "!eval" {
			if len(s) == 0 {
				return nil, fmt.Errorf("eval requires at least one argument")
			}
			return eval(s[0].(string), s[1:])
		}

		return s, nil
	case yaml.MappingNode:
		m := make(map[string]any)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			key := keyNode.Value
			value, err := bake(valueNode)
			if err != nil {
				return nil, err
			}
			m[key] = value
		}
		return m, nil
	case yaml.ScalarNode:
		return node.Value, nil
	case yaml.AliasNode:
		return bake(node.Alias)
	}

	return nil, fmt.Errorf("unsupported node kind: %v", node.Kind)
}

func main() {

	/*
	data :=`
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
  namespace: default
  labels:
    app: &appname example-app
    name: *appname
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config2
  namespace: default
`
*/

data := `
mystring: !eval
  - join
  - hello
  - ' '
  - world
`

/*
data := `
!eval
- join
- hello
- ' '
- world
`
*/

/*
	fmt.Printf("DocumentNode: %v\n", yaml.DocumentNode) // 1
	fmt.Printf("SequenceNode: %v\n", yaml.SequenceNode) // 2
	fmt.Printf("MappingNode: %v\n", yaml.MappingNode) // 4
	fmt.Printf("ScalarNode: %v\n", yaml.ScalarNode) // 8
	fmt.Printf("AliasNode: %v\n", yaml.AliasNode) // 16
*/

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

		/*
		jsonData, err := json.MarshalIndent(root, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(jsonData))
		*/
		baked, err := bake(&root)
		if err != nil {
			panic(err)
		}
		bakedJSON, err := json.MarshalIndent(baked, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bakedJSON))
	}
}


