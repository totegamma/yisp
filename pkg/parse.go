package yisp

import (
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"github.com/totegamma/yisp/internal/yaml"
	"io"
	"reflect"
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

func ParseAny(filename string, v any) (*YispNode, error) {

	typ := reflect.TypeOf(v)

	switch typ.Kind() {
	case reflect.Bool:
		return &YispNode{
			Kind:  KindBool,
			Value: v,
			Pos: Position{
				File: filename,
			},
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &YispNode{
			Kind:  KindInt,
			Value: v,
			Pos: Position{
				File: filename,
			},
		}, nil
	case reflect.Float32, reflect.Float64:
		return &YispNode{
			Kind:  KindFloat,
			Value: v,
			Pos: Position{
				File: filename,
			},
		}, nil
	case reflect.String:
		return &YispNode{
			Kind:  KindString,
			Value: v,
			Pos: Position{
				File: filename,
			},
		}, nil
	case reflect.Map:
		m := NewYispMap()
		dict, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map[string]any, got %T", v)
		}

		for key, value := range dict {
			itemNode, err := ParseAny(filename, value)
			if err != nil {
				return nil, err
			}
			m.Set(key, itemNode)
		}

		return &YispNode{
			Kind:  KindMap,
			Value: m,
			Pos: Position{
				File: filename,
			},
		}, nil
	case reflect.Slice, reflect.Array:
		val, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("expected []any, got %T", v)
		}
		s := make([]any, len(val))

		for i, item := range val {
			itemNode, err := ParseAny(filename, item)
			if err != nil {
				return nil, err
			}
			s[i] = itemNode
		}

		return &YispNode{
			Kind:  KindArray,
			Value: s,
			Pos: Position{
				File: filename,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", typ.Kind())
	}
}

func ParseJson(filename string, reader io.Reader) (*YispNode, error) {
	var data any
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	node, err := ParseAny(filename, data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return node, nil
}
