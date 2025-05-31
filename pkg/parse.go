package yisp

import (
	"github.com/rs/xid"
	"github.com/totegamma/yisp/internal/yaml"
)

// Parse converts a YAML node to a YispNode
func Parse(filename string, node *yaml.Node) (*YispNode, error) {
	var result *YispNode
	var err error

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		result, err = Parse(filename, node.Content[0])

	case yaml.SequenceNode:
		s := make([]any, len(node.Content))
		for i, item := range node.Content {
			value, err := Parse(filename, item)
			if err != nil {
				return nil, err
			}
			s[i] = value
		}

		result = &YispNode{
			Kind:  KindArray,
			Value: s,
			Tag:   node.Tag,
			Pos: Position{
				File:   filename,
				Line:   node.Line,
				Column: node.Column,
			},
		}

	case yaml.MappingNode:
		m := NewYispMap()
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			key := keyNode.Value
			if key == "<<" {
				key = YISP_SPECIAL_MERGE_KEY + xid.New().String()
			}

			value, err := Parse(filename, valueNode)
			if err != nil {
				return nil, err
			}
			m.Set(key, value)
		}

		result = &YispNode{
			Kind:  KindMap,
			Value: m,
			Tag:   node.Tag,
			Pos: Position{
				File:   filename,
				Line:   node.Line,
				Column: node.Column,
			},
		}

	case yaml.ScalarNode:

		kind := KindString
		switch node.Tag {
		case "!null", "!!null":
			kind = KindNull
		case "!bool", "!!bool":
			kind = KindBool
		case "!int", "!!int":
			kind = KindInt
		case "!float", "!!float":
			kind = KindFloat
		case "!string", "!!str":
			kind = KindString
		}

		result = &YispNode{
			Kind:  kind,
			Value: node.Value,
			Tag:   node.Tag,
			Pos: Position{
				File:   filename,
				Line:   node.Line,
				Column: node.Column,
			},
		}

	case yaml.AliasNode:
		result = &YispNode{
			Kind:  KindSymbol,
			Value: node.Value,
			Tag:   node.Tag,
			Pos: Position{
				File:   filename,
				Line:   node.Line,
				Column: node.Column,
			},
		}
	}

	if node.Anchor != "" {
		result.Anchor = node.Anchor
	}

	return result, err
}
