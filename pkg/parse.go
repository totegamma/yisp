package yisp

import (
	"github.com/totegamma/yisp/yaml"
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
			Kind:   KindArray,
			Value:  s,
			Tag:    node.Tag,
			File:   filename,
			Line:   node.Line,
			Column: node.Column,
		}

	case yaml.MappingNode:
		m := make(map[string]any)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			key := keyNode.Value
			value, err := Parse(filename, valueNode)
			if err != nil {
				return nil, err
			}
			m[key] = value
		}

		result = &YispNode{
			Kind:   KindMap,
			Value:  m,
			Tag:    node.Tag,
			File:   filename,
			Line:   node.Line,
			Column: node.Column,
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
			Kind:   kind,
			Value:  node.Value,
			Tag:    node.Tag,
			File:   filename,
			Line:   node.Line,
			Column: node.Column,
		}

	case yaml.AliasNode:
		result = &YispNode{
			Kind:   KindSymbol,
			Value:  node.Value,
			Tag:    node.Tag,
			File:   filename,
			Line:   node.Line,
			Column: node.Column,
		}
	}

	if node.Anchor != "" {
		result.Anchor = node.Anchor
	}

	return result, err
}
