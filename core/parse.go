package core

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

func ParseAny(filename string, v any) (*YispNode, error) {

	typ := reflect.TypeOf(v)

	switch typ.Kind() {
	case reflect.Bool:
		return &YispNode{
			Kind:  KindBool,
			Value: v,
			Attr: Attribute{
				File: filename,
			},
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &YispNode{
			Kind:  KindInt,
			Value: v,
			Attr: Attribute{
				File: filename,
			},
		}, nil
	case reflect.Float32, reflect.Float64:
		return &YispNode{
			Kind:  KindFloat,
			Value: v,
			Attr: Attribute{
				File: filename,
			},
		}, nil
	case reflect.String:
		return &YispNode{
			Kind:  KindString,
			Value: v,
			Attr: Attribute{
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
			Attr: Attribute{
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
			Attr: Attribute{
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
