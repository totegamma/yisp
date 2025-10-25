package engine

import (
	"github.com/rs/xid"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/internal/yaml"
)

// Parse converts a YAML node to a core.YispNode
func Parse(filename string, node *yaml.Node) (*core.YispNode, error) {
	var result *core.YispNode
	var err error

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		result, err = Parse(filename, node.Content[0])
		result.Attr = core.Attribute{
			Sources: []core.FilePos{
				{
					File:   filename,
					Line:   node.Line,
					Column: node.Column,
				},
			},
			HeadComment: node.HeadComment,
			LineComment: node.LineComment,
			FootComment: node.FootComment,
			Style:       node.Style,
		}

	case yaml.SequenceNode:
		s := make([]any, len(node.Content))
		for i, item := range node.Content {
			value, err := Parse(filename, item)
			if err != nil {
				return nil, err
			}
			s[i] = value
		}

		result = &core.YispNode{
			Kind:  core.KindArray,
			Value: s,
			Tag:   node.Tag,
			Attr: core.Attribute{
				Sources: []core.FilePos{
					{
						File:   filename,
						Line:   node.Line,
						Column: node.Column,
					},
				},
				HeadComment: node.HeadComment,
				LineComment: node.LineComment,
				FootComment: node.FootComment,
				Style:       node.Style,
			},
		}

	case yaml.MappingNode:
		m := core.NewYispMap()
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			key := keyNode.Value
			if key == "<<" {
				key = core.YISP_SPECIAL_MERGE_KEY + xid.New().String()
			}

			value, err := Parse(filename, valueNode)
			if err != nil {
				return nil, err
			}

			value.Attr.KeyStyle = keyNode.Style
			value.Attr.KeyHeadComment = keyNode.HeadComment
			value.Attr.KeyLineComment = keyNode.LineComment
			value.Attr.KeyFootComment = keyNode.FootComment

			m.Set(key, value)
		}

		result = &core.YispNode{
			Kind:  core.KindMap,
			Value: m,
			Tag:   node.Tag,
			Attr: core.Attribute{
				Sources: []core.FilePos{
					{
						File:   filename,
						Line:   node.Line,
						Column: node.Column,
					},
				},
				HeadComment: node.HeadComment,
				LineComment: node.LineComment,
				FootComment: node.FootComment,
				Style:       node.Style,
			},
		}

	case yaml.ScalarNode:

		kind := core.KindString
		switch node.Tag {
		case "!null", "!!null":
			kind = core.KindNull
		case "!bool", "!!bool":
			kind = core.KindBool
		case "!int", "!!int":
			kind = core.KindInt
		case "!float", "!!float":
			kind = core.KindFloat
		case "!string", "!!str":
			kind = core.KindString
		}

		result = &core.YispNode{
			Kind:  kind,
			Value: node.Value,
			Tag:   node.Tag,
			Attr: core.Attribute{
				Sources: []core.FilePos{
					{
						File:   filename,
						Line:   node.Line,
						Column: node.Column,
					},
				},
				HeadComment: node.HeadComment,
				LineComment: node.LineComment,
				FootComment: node.FootComment,
				Style:       node.Style,
			},
		}

	case yaml.AliasNode:
		result = &core.YispNode{
			Kind:  core.KindSymbol,
			Value: node.Value,
			Tag:   node.Tag,
			Attr: core.Attribute{
				Sources: []core.FilePos{
					{
						File:   filename,
						Line:   node.Line,
						Column: node.Column,
					},
				},
				HeadComment: node.HeadComment,
				LineComment: node.LineComment,
				FootComment: node.FootComment,
				Style:       node.Style,
			},
		}
	}

	if node.Anchor != "" {
		result.Anchor = node.Anchor
	}

	return result, err
}
