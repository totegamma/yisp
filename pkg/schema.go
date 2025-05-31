package yisp

import (
	"fmt"
	"slices"
)

type Schema struct {
	Type                 string             `json:"type"`
	Required             []string           `json:"required,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties bool               `json:"additionalProperties,omitempty"`
	Arguments            []Schema           `json:"arguments,omitempty"`
	Returns              *Schema            `json:"returns,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              any                `json:"default,omitempty"`
	PatchStrategy        string             `json:"patchStrategy,omitempty"`
	PatchMergeKey        string             `json:"patchMergeKey,omitempty"`
}

func (s *Schema) Validate(node *YispNode) error {

	switch s.Type {
	case "any":
		return nil
	case "null":
		if node.Kind != KindNull {
			return fmt.Errorf("expected null, got %s", node.Kind)
		}
	case "boolean":
		if node.Kind != KindBool {
			return fmt.Errorf("expected bool, got %s", node.Kind)
		}
	case "integer":
		if node.Kind != KindInt {
			return fmt.Errorf("expected int, got %s", node.Kind)
		}
	case "float":
		if node.Kind != KindFloat {
			return fmt.Errorf("expected float, got %s", node.Kind)
		}
	case "string":
		if node.Kind != KindString {
			return fmt.Errorf("expected string, got %s", node.Kind)
		}
	case "array":
		if node.Kind != KindArray {
			return fmt.Errorf("expected array, got %s", node.Kind)
		}
		if s.Items != nil {
			arr, ok := node.Value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", node.Value)
			}
			for _, item := range arr {
				itemNode, ok := item.(*YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", item)
				}
				if err := s.Items.Validate(itemNode); err != nil {
					return err
				}
			}
		}
	case "object":
		if node.Kind != KindMap {
			return fmt.Errorf("expected map, got %s", node.Kind)
		}
		m, ok := node.Value.(*YispMap)
		if !ok {
			return fmt.Errorf("expected map, got %T", node.Value)
		}

		processed := make(map[string]bool)
		for key, subSchema := range s.Properties {
			item, ok := m.Get(key)
			if !ok {
				if slices.Contains(s.Required, key) {
					return fmt.Errorf("missing required property: %s", key)
				}
			}
			itemNode, ok := item.(*YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", itemNode)
			}
			if err := subSchema.Validate(itemNode); err != nil {
				return err
			}
			processed[key] = true
		}

		for key := range m.AllFromFront() {
			if _, ok := processed[key]; !ok {
				if !s.AdditionalProperties {
					return fmt.Errorf("unexpected property: %s", key)
				}
			}
		}

	case "function":
		if node.Kind != KindLambda {
			return fmt.Errorf("expected function, got %s", node.Kind)
		}
		fn, ok := node.Value.(*Lambda)
		if !ok {
			return fmt.Errorf("expected YispLambda, got %T", node.Value)
		}
		if len(fn.Arguments) != len(s.Arguments) {
			return fmt.Errorf("expected %d arguments, got %d", len(s.Arguments), len(fn.Arguments))
		}
		for i, arg := range s.Arguments {
			if fn.Arguments[i].Schema != nil && !arg.Equals(fn.Arguments[i].Schema) {
				return fmt.Errorf("argument %d does not match schema", i)
			}

		}
		if s.Returns != nil && fn.Returns != nil {
			if !s.Returns.Equals(fn.Returns) {
				return fmt.Errorf("return type does not match schema. Expected %s, got %s", s.Returns.Type, fn.Returns.Type)
			}
		}

	default:
		return fmt.Errorf("unknown type: %s", s.Type)
	}

	return nil
}

func (s *Schema) Cast(node *YispNode) {
}

func (s *Schema) Equals(other *Schema) bool {

	if s.Type != other.Type {
		return false
	}

	switch s.Type {
	case "any", "null", "boolean", "integer", "float", "string":
		return true
	case "array":
		if s.Items == nil && other.Items == nil {
			return true
		}
		if s.Items == nil || other.Items == nil {
			return false
		}
		return s.Items.Equals(other.Items)
	case "object":
		if len(s.Properties) != len(other.Properties) {
			return false
		}
		for key, subSchema := range s.Properties {
			otherSubSchema, ok := other.Properties[key]
			if !ok || !subSchema.Equals(otherSubSchema) {
				return false
			}
		}
		for key := range s.Properties {
			if _, ok := other.Properties[key]; !ok {
				return false
			}
		}
		return true
	case "function":
		if len(s.Arguments) != len(other.Arguments) {
			return false
		}
		for i, arg := range s.Arguments {
			if !arg.Equals(&other.Arguments[i]) {
				return false
			}
		}
		if s.Returns == nil && other.Returns == nil {
			return true
		}
		if s.Returns == nil || other.Returns == nil {
			return false
		}
		return s.Returns.Equals(other.Returns)
	default:
		return false
	}
}
