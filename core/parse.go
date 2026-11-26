package core

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"gopkg.in/yaml.v3"
)

func ParseAny(filename string, v any) (*YispNode, error) {

	typ := reflect.TypeOf(v)
	if typ == nil {
		return &YispNode{
			Kind:  KindNull,
			Value: nil,
			Attr: Attribute{
				Sources: []FilePos{
					{
						File: filename,
					},
				},
			},
		}, nil
	}

	switch typ.Kind() {
	case reflect.Bool:
		return &YispNode{
			Kind:  KindBool,
			Value: v,
			Attr: Attribute{
				Sources: []FilePos{
					{
						File: filename,
					},
				},
			},
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &YispNode{
			Kind:  KindInt,
			Value: v,
			Attr: Attribute{
				Sources: []FilePos{
					{
						File: filename,
					},
				},
			},
		}, nil
	case reflect.Float32, reflect.Float64:
		return &YispNode{
			Kind:  KindFloat,
			Value: v,
			Attr: Attribute{
				Sources: []FilePos{
					{
						File: filename,
					},
				},
			},
		}, nil
	case reflect.String:
		return &YispNode{
			Kind:  KindString,
			Value: v,
			Attr: Attribute{
				Sources: []FilePos{
					{
						File: filename,
					},
				},
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
			Attr: Attribute{
				Sources: []FilePos{
					{
						File: filename,
					},
				},
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
			Attr: Attribute{
				Sources: []FilePos{
					{
						File: filename,
					},
				},
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

// ParseYaml parses standard YAML (not yisp-extended YAML) into a YispNode
func ParseYaml(filename string, reader io.Reader) (*YispNode, error) {
	var doc yaml.Node
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %v", err)
	}

	node, err := parseYamlNode(filename, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return node, nil
}

func parseYamlNode(filename string, node *yaml.Node) (*YispNode, error) {
	pos := Attribute{
		Sources: []FilePos{
			{
				File:   filename,
				Line:   node.Line,
				Column: node.Column,
			},
		},
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return ParseAny(filename, nil)
		}
		return parseYamlNode(filename, node.Content[0])

	case yaml.AliasNode:
		return parseYamlNode(filename, node.Alias)

	case yaml.SequenceNode:
		s := make([]any, len(node.Content))
		for i, item := range node.Content {
			itemNode, err := parseYamlNode(filename, item)
			if err != nil {
				return nil, err
			}
			s[i] = itemNode
		}
		return &YispNode{
			Kind:  KindArray,
			Value: s,
			Attr:  pos,
		}, nil

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == "<<" && node.Content[i].Tag == "!!merge" {
				// fall back to the generic decoder to resolve merge keys
				var data any
				if err := node.Decode(&data); err != nil {
					return nil, err
				}
				result, err := ParseAny(filename, data)
				if err != nil {
					return nil, err
				}
				result.Attr = pos
				return result, nil
			}
		}

		m := NewYispMap()
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode, err := parseYamlNode(filename, node.Content[i+1])
			if err != nil {
				return nil, err
			}
			valueNode.Attr.KeyPos = &FilePos{
				File:   filename,
				Line:   keyNode.Line,
				Column: keyNode.Column,
			}
			m.Set(keyNode.Value, valueNode)
		}
		return &YispNode{
			Kind:  KindMap,
			Value: m,
			Attr:  pos,
		}, nil

	case yaml.ScalarNode:
		var data any
		if err := node.Decode(&data); err != nil {
			return nil, err
		}
		result, err := ParseAny(filename, data)
		if err != nil {
			return nil, err
		}
		result.Attr = pos
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported YAML node kind: %v", node.Kind)
	}
}
