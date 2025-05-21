package yisp

import (
	"bytes"
	"fmt"
	"github.com/totegamma/yisp/yaml"
)

// Render converts a YispNode to a native Go value
func render(node *YispNode) (*yaml.Node, error) {
	switch node.Kind {
	case KindNull:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "null",
		}, nil

	case KindBool, KindInt, KindFloat, KindString:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: fmt.Sprintf("%v", node.Value),
		}, nil
	case KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array value. Actual type: %T", node.Value)
		}
		results := make([]*yaml.Node, len(arr))
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
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Content: results,
		}, nil
	case KindMap:
		m, ok := node.Value.(*YispMap)
		if !ok {
			return nil, fmt.Errorf("invalid map value")
		}
		results := make([]*yaml.Node, 0)
		for key, item := range m.AllFromFront() {
			node, ok := item.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}

			content, err := render(node)
			if err != nil {
				return nil, err
			}

			results = append(results, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: key,
			})

			results = append(results, content)

		}
		return &yaml.Node{
			Kind:    yaml.MappingNode,
			Content: results,
		}, nil

	case KindLambda:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "(lambda)",
		}, nil
	case KindParameter:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "(parameter)",
		}, nil
	case KindSymbol:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "(symbol)",
		}, nil
	case KindType:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "(type)",
		}, nil
	default:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "(unknown)",
		}, nil
	}
}

func Render(node *YispNode) (string, error) {

	if node.Kind == KindArray {
		arr, ok := node.Value.([]any)
		if !ok {
			return "", fmt.Errorf("invalid array value(root)")
		}

		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		for _, item := range arr {
			node, ok := item.(*YispNode)
			if !ok {
				return "", fmt.Errorf("invalid item type: %T", item)
			}
			rendered, err := render(node)
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
