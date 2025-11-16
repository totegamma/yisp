package engine

import (
	"bytes"
	"fmt"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/internal/yaml"
)

func (e *engine) getLineComment(node *core.YispNode) string {
	lineComment := node.Attr.LineComment
	if e.renderSources {
		lineComment += node.Sourcemap()
	}
	return lineComment
}

// Render converts a YispNode to a yaml.Node
func (e *engine) renderYamlNodes(node *core.YispNode) (*yaml.Node, error) {
	switch node.Kind {
	case core.KindNull:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       "null",
			Tag:         "!!null",
			HeadComment: node.Attr.HeadComment,
			LineComment: e.getLineComment(node),
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindBool:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%t", node.Value),
			Tag:         "!!bool",
			HeadComment: node.Attr.HeadComment,
			LineComment: e.getLineComment(node),
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindInt:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%d", node.Value),
			Tag:         "!!int",
			HeadComment: node.Attr.HeadComment,
			LineComment: e.getLineComment(node),
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindFloat:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%f", node.Value),
			Tag:         "!!float",
			HeadComment: node.Attr.HeadComment,
			LineComment: e.getLineComment(node),
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindString:
		return &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fmt.Sprintf("%s", node.Value),
			Tag:         "!!str",
			HeadComment: node.Attr.HeadComment,
			LineComment: e.getLineComment(node),
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
			results[i], err = e.renderYamlNodes(node)
			if err != nil {
				return nil, err
			}
		}
		return &yaml.Node{
			Kind:        yaml.SequenceNode,
			Content:     results,
			HeadComment: node.Attr.HeadComment,
			LineComment: e.getLineComment(node),
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

			content, err := e.renderYamlNodes(node)
			if err != nil {
				return nil, err
			}

			results = append(results, &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       key,
				HeadComment: node.Attr.KeyHeadComment,
				LineComment: e.getLineComment(node),
				FootComment: node.Attr.KeyFootComment,
				Style:       node.Attr.KeyStyle,
			})

			results = append(results, content)

		}

		return &yaml.Node{
			Kind:        yaml.MappingNode,
			Content:     results,
			HeadComment: node.Attr.HeadComment,
			LineComment: e.getLineComment(node),
			FootComment: node.Attr.FootComment,
			Style:       node.Attr.Style,
		}, nil

	case core.KindLambda:
		if e.renderSpecialObjects {
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
				LineComment: e.getLineComment(node),
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	case core.KindParameter:
		if e.renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(parameter)",
				HeadComment: node.Attr.HeadComment,
				LineComment: e.getLineComment(node),
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	case core.KindSymbol:
		if e.renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(symbol)",
				HeadComment: node.Attr.HeadComment,
				LineComment: e.getLineComment(node),
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	case core.KindType:
		if e.renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(type)",
				HeadComment: node.Attr.HeadComment,
				LineComment: e.getLineComment(node),
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	default:
		if e.renderSpecialObjects {
			return &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       "(unknown)",
				HeadComment: node.Attr.HeadComment,
				LineComment: e.getLineComment(node),
				FootComment: node.Attr.FootComment,
				Style:       node.Attr.Style,
			}, nil
		}
	}

	return nil, nil
}

func (e *engine) Render(node *core.YispNode) (string, error) {

	// Verify types on the whole object tree before rendering
	if err := core.VerifyTypes(node, e.allowUntypedManifest); err != nil {
		return "", fmt.Errorf("type verification failed: %v", err)
	}

	if node.Kind == core.KindArray && node.IsDocumentRoot {
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
			rendered, err := e.renderYamlNodes(node)
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
		rendered, err := e.renderYamlNodes(node)
		if err != nil {
			return "", err
		}
		b, _ := yaml.Marshal(rendered)
		return string(b), nil
	}
}
