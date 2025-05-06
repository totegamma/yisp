package yisp

import (
	"bytes"
	"fmt"
	"github.com/totegamma/yisp/yaml"
)

// Render converts a YispNode to a native Go value
func render(node *YispNode) (any, error) {
	switch node.Kind {
	case KindNull, KindBool, KindInt, KindFloat, KindString:
		return node.Value, nil
	case KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array value. Actual type: %T", node.Value)
		}
		results := make([]any, len(arr))
		for i, item := range arr {
			node, ok := item.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}
			var err error
			results[i], err = render(node)
			if err != nil {
				return nil, err
			}
		}
		return results, nil
	case KindMap:
		m, ok := node.Value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid map value")
		}
		results := make(map[string]any)
		for key, item := range m {
			node, ok := item.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}
			var err error
			results[key], err = render(node)
			if err != nil {
				return nil, err
			}
		}
		return results, nil
	case KindLambda:
		return "(lambda)", nil
	case KindParameter:
		return "(parameter)", nil
	case KindSymbol:
		return "(symbol)", nil
	default:
		return "(unknown)", nil
	}
}

func Flatten(node *YispNode) (*YispNode, error) {
	if node.Kind == KindArray && node.Tag == "!expand" {
		arr, ok := node.Value.([]*YispNode)
		if !ok {
			return nil, fmt.Errorf("invalid array value")
		}

		results := make([]*YispNode, 0)
		for _, item := range arr {
			flattened, err := Flatten(item)
			if err != nil {
				return nil, err
			}

			if flattened.Kind == KindArray {
				// Flatten the inner array
				innerArr, ok := flattened.Value.([]*YispNode)
				if !ok {
					return nil, fmt.Errorf("invalid inner array value")
				}
				results = append(results, innerArr...)
			} else {
				results = append(results, flattened)
			}
		}
		return &YispNode{
			Kind:  KindArray,
			Value: results,
			Tag:   node.Tag,
		}, nil
	}

	return node, nil
}

func Render(node *YispNode) (string, error) {

	node.Tag = "!expand"

	var err error
	node, err = Flatten(node)
	if err != nil {
		return "", err
	}

	if node.Kind == KindArray {
		arr, ok := node.Value.([]*YispNode)
		if !ok {
			return "", fmt.Errorf("invalid array value(root)")
		}

		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		for _, item := range arr {
			rendered, err := render(item)
			if err != nil {
				return "", err
			}

			if rendered == nil {
				continue
			}

			enc.Encode(rendered)
		}
		enc.Close()
		return buf.String(), nil

	} else {
		rendered, err := render(node)
		if err != nil {
			return "", err
		}
		b, _ := yaml.Marshal(rendered)
		return string(b), nil
	}
}
