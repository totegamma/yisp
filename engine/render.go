package engine

import (
	"bytes"
	"fmt"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/internal/yaml"
)

// Render converts a YispNode to a native Go value
func render(node *core.YispNode, renderSpecialObjects bool) (*yaml.Node, error) {
	switch node.Kind {
	case core.KindNull:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       "null",
			Tag:         "!!null",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindBool:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%t", node.Value),
			Tag:         "!!bool",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindInt:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%d", node.Value),
			Tag:         "!!int",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindFloat:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%f", node.Value),
			Tag:         "!!float",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindString:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%s", node.Value),
			Tag:         "!!str",
			HeadComment: node.Attr.HeadComment,
			LineComment: node.Attr.LineComment,
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array value. Actual type: %T", node.Value)
		}
		results := make([]*yaml.Node, len(arr))
		for i, item := range arr {
			node, ok := item.(*core.YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}
			var err error
			results[i], err = render(node, renderSpecialObjects)
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
	case core.KindMap:
		m, ok := node.Value.(*core.YispMap)
		if !ok {
			return nil, fmt.Errorf("invalid map value")
		}
		results := make([]*yaml.Node, 0)
		for key, item := range m.AllFromFront() {
			node, ok := item.(*core.YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}

			content, err := render(node, renderSpecialObjects)
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

	case core.KindLambda:
		if renderSpecialObjects {
			value := "λ"
			lambda := node.Value.(*core.Lambda)
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
	case core.KindParameter:
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
	case core.KindSymbol:
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
	case core.KindType:
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

func (e *engine) Render(node *core.YispNode) (string, error) {

	if node.Kind == core.KindArray {
		arr, ok := node.Value.([]any)
		if !ok {
			return "", fmt.Errorf("invalid array value(root)")
		}

		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		for _, item := range arr {
			node, ok := item.(*core.YispNode)
			if !ok {
				return "", fmt.Errorf("invalid item type: %T", item)
			}
			rendered, err := render(node, e.renderSpecialObjects)
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
		rendered, err := render(node, e.renderSpecialObjects)
		if err != nil {
			return "", err
		}
		b, _ := yaml.Marshal(rendered)
		return string(b), nil
	}
}
