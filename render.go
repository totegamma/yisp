package yisp

import (
	"bytes"
	"fmt"
	"github.com/totegamma/yisp/yaml"
)

// Render converts a YispNode to a native Go value
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

func Render(node *YispNode) (string, error) {
	if node.Kind == KindArray {
		arr, ok := node.Value.([]*YispNode)
		if !ok {
			return "", fmt.Errorf("invalid array value")
		}

		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		for _, item := range arr {
			rendered := render(item)
			enc.Encode(rendered)
		}
		enc.Close()
		return buf.String(), nil

	} else {
		rendered := render(node)
		b, _ := yaml.Marshal(rendered)
		return string(b), nil
	}
}
