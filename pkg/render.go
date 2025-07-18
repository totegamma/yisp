package yisp

import (
	"bytes"
	"fmt"

	"github.com/totegamma/yisp/internal/yaml"
)

// Render converts a YispNode to a native Go value
func render(node *YispNode) (*yaml.Node, error) {
	switch node.Kind {
	case KindNull:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       "null",
			Tag:         "!!null",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case KindBool:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%t", node.Value),
			Tag:         "!!bool",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case KindInt:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%d", node.Value),
			Tag:         "!!int",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case KindFloat:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%f", node.Value),
			Tag:         "!!float",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case KindString:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%s", node.Value),
			Tag:         "!!str",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
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
			Kind:        yaml.SequenceNode,
			Content:     results,
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
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
				Kind:        yaml.ScalarNode,
				Value:       key,
				HeadComment: node.Attr.KeyHeadComment,
				LineComment: node.Attr.KeyLineComment,
				FootComment: node.Attr.KeyFootComment,
				Style:       node.Attr.KeyStyle,
			})

			results = append(results, content)

		}

		return &yaml.Node{
			Kind:        yaml.MappingNode,
			Content:     results,
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case KindLambda:
		if renderSpecialObjects {
			value := "λ"
			lambda := node.Value.(*Lambda)
			for i, arg := range lambda.Arguments {
				value += arg.Name
				if i < len(lambda.Arguments)-1 {
					value += ","
				}
			}
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       value,
				HeadComment: node.Attr.HeadComment,
				LineComment: node.Attr.LineComment,
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	case KindParameter:
		if renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(parameter)",
				HeadComment: node.Attr.HeadComment,
				LineComment: node.Attr.LineComment,
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	case KindSymbol:
		if renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(symbol)",
				HeadComment: node.Attr.HeadComment,
				LineComment: node.Attr.LineComment,
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	case KindType:
		if renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(type)",
				HeadComment: node.Attr.HeadComment,
				LineComment: node.Attr.LineComment,
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	default:
		if renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(unknown)",
				HeadComment: node.Attr.HeadComment,
				LineComment: node.Attr.LineComment,
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	}

	return nil, nil
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
