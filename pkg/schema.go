package yisp

import (
	"fmt"
	"slices"
	"strconv"
)

var schemaTypeToKind = map[string]Kind{
	"null":     KindNull,
	"boolean":  KindBool,
	"integer":  KindInt,
	"float":    KindFloat,
	"string":   KindString,
	"array":    KindArray,
	"object":   KindMap,
	"function": KindLambda,
}

type Schema struct {
	Type                 string             `json:"type"`
	Required             []string           `json:"required,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties bool               `json:"additionalProperties,omitempty"`
	Arguments            []*Schema          `json:"arguments,omitempty"`
	Returns              *Schema            `json:"returns,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              any                `json:"default,omitempty"`
	PatchStrategy        string             `json:"patchStrategy,omitempty"`
	PatchMergeKey        string             `json:"patchMergeKey,omitempty"`
	OneOf                []*Schema          `json:"oneOf,omitempty"`

	// Numeric constraints
	MultipleOf       *int     `json:"multipleOf,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`

	// String constraints
	MinLength *int `json:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty"`
}

func (s *Schema) Validate(node *YispNode) error {

	if s.OneOf != nil {
		var errors []string
		for _, subSchema := range s.OneOf {
			err := subSchema.Validate(node)
			if err == nil {
				return nil // Valid against one of the schemas
			}
			errors = append(errors, err.Error())
		}
		return fmt.Errorf("node does not match any of the oneOf schemas: %s", errors)
	}

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
		if s.Minimum != nil && node.Value.(int) < int(*s.Minimum) {
			return fmt.Errorf("value %d is less than minimum %f", node.Value.(int), *s.Minimum)
		}
		if s.Maximum != nil && node.Value.(int) > int(*s.Maximum) {
			return fmt.Errorf("value %d is greater than maximum %f", node.Value.(int), *s.Maximum)
		}
		if s.ExclusiveMinimum != nil && node.Value.(int) <= int(*s.ExclusiveMinimum) {
			return fmt.Errorf("value %d is not greater than exclusive minimum %f", node.Value.(int), *s.ExclusiveMinimum)
		}
		if s.ExclusiveMaximum != nil && node.Value.(int) >= int(*s.ExclusiveMaximum) {
			return fmt.Errorf("value %d is not less than exclusive maximum %f", node.Value.(int), *s.ExclusiveMaximum)
		}
		if s.MultipleOf != nil {
			if node.Value.(int)%*s.MultipleOf != 0 {
				return fmt.Errorf("value %d is not a multiple of %d", node.Value.(int), *s.MultipleOf)
			}
		}
	case "float":
		if node.Kind != KindFloat {
			return fmt.Errorf("expected float, got %s", node.Kind)
		}
		if s.Minimum != nil && node.Value.(float64) < *s.Minimum {
			return fmt.Errorf("value %f is less than minimum %f", node.Value.(float64), *s.Minimum)
		}
		if s.Maximum != nil && node.Value.(float64) > *s.Maximum {
			return fmt.Errorf("value %f is greater than maximum %f", node.Value.(float64), *s.Maximum)
		}
		if s.ExclusiveMinimum != nil && node.Value.(float64) <= *s.ExclusiveMinimum {
			return fmt.Errorf("value %f is not greater than exclusive minimum %f", node.Value.(float64), *s.ExclusiveMinimum)
		}
		if s.ExclusiveMaximum != nil && node.Value.(float64) >= *s.ExclusiveMaximum {
			return fmt.Errorf("value %f is not less than exclusive maximum %f", node.Value.(float64), *s.ExclusiveMaximum)
		}
		if s.MultipleOf != nil {
			if int(node.Value.(float64))%*s.MultipleOf != 0 {
				return fmt.Errorf("value %f is not a multiple of %d", node.Value.(float64), *s.MultipleOf)
			}
		}
	case "string":
		if node.Kind != KindString {
			return fmt.Errorf("expected string, got %s", node.Kind)
		}
		if s.MinLength != nil && len(node.Value.(string)) < *s.MinLength {
			return fmt.Errorf("string length %d is less than minimum %d", len(node.Value.(string)), *s.MinLength)
		}
		if s.MaxLength != nil && len(node.Value.(string)) > *s.MaxLength {
			return fmt.Errorf("string length %d is greater than maximum %d", len(node.Value.(string)), *s.MaxLength)
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
				} else {
					continue
				}
			}
			itemNode, ok := item.(*YispNode)
			if !ok {
				return fmt.Errorf("[object]expected YispNode, got %T", item)
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

func (s *Schema) Cast(node *YispNode) (*YispNode, error) {
	switch s.Type {
	case "any":
		return node, nil
	case "null":
		return &YispNode{
			Kind:  KindNull,
			Value: nil,
			Tag:   node.Tag,
			Pos:   node.Pos,
			Type:  s,
		}, nil
	case "boolean":
		value, err := isTruthy(node)
		if err != nil {
			return nil, err
		}

		return &YispNode{
			Kind:  KindBool,
			Value: value,
			Tag:   node.Tag,
			Pos:   node.Pos,
			Type:  s,
		}, nil

	case "integer":
		value := 0
		if node.Kind == KindInt {
			return node, nil
		} else if node.Kind == KindFloat {
			value = int(node.Value.(float64)) // Convert float to int
		} else if node.Kind == KindString {
			var err error
			value, err = strconv.Atoi(node.Value.(string))
			if err != nil {
				return nil, fmt.Errorf("cannot cast string to int: %v", err)
			}
		} else if node.Kind == KindBool {
			if node.Value == true {
				value = 1
			} else {
				value = 0
			}
		} else {
			return nil, fmt.Errorf("cannot cast %s to integer", node.Kind)
		}
		return &YispNode{
			Kind:  KindInt,
			Value: value,
			Tag:   node.Tag,
			Pos:   node.Pos,
			Type:  s,
		}, nil

	case "float":
		value := 0.0
		if node.Kind == KindFloat {
			return node, nil
		} else if node.Kind == KindInt {
			value = float64(node.Value.(int)) // Convert int to float
		} else if node.Kind == KindString {
			var err error
			value, err = strconv.ParseFloat(node.Value.(string), 64)
			if err != nil {
				return nil, fmt.Errorf("cannot cast string to float: %v", err)
			}
		} else if node.Kind == KindBool {
			if node.Value == true {
				value = 1.0
			} else {
				value = 0.0
			}
		} else {
			return nil, fmt.Errorf("cannot cast %s to float", node.Kind)
		}

		return &YispNode{
			Kind:  KindFloat,
			Value: value,
			Tag:   node.Tag,
			Pos:   node.Pos,
			Type:  s,
		}, nil

	case "string":
		value := ""
		if node.Kind == KindString {
			return node, nil
		} else if node.Kind == KindInt {
			value = strconv.Itoa(node.Value.(int)) // Convert int to string
		} else if node.Kind == KindFloat {
			value = strconv.FormatFloat(node.Value.(float64), 'f', -1, 64) // Convert float to string
		} else if node.Kind == KindBool {
			if node.Value == true {
				value = "true"
			} else {
				value = "false"
			}
		} else if node.Kind == KindNull {
			value = "null"
		} else {
			return nil, fmt.Errorf("cannot cast %s to string", node.Kind)
		}
		return &YispNode{
			Kind:  KindString,
			Value: value,
			Tag:   node.Tag,
			Pos:   node.Pos,
			Type:  s,
		}, nil

	case "array":
		if node.Kind != KindArray {
			return nil, fmt.Errorf("expected array, got %s", node.Kind)
		}
		if s.Items == nil {
			return node, nil // No item schema, return as is
		}
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("expected array, got %T", node.Value)
		}
		newArr := make([]any, len(arr))
		for i, item := range arr {
			itemNode, ok := item.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("expected YispNode, got %T", item)
			}
			castedNode, err := s.Items.Cast(itemNode)
			if err != nil {
				return nil, fmt.Errorf("error casting array item at index %d: %v", i, err)
			}
			newArr[i] = castedNode
		}
		return &YispNode{
			Kind:  KindArray,
			Value: newArr,
			Tag:   node.Tag,
			Pos:   node.Pos,
			Type:  s,
		}, nil

	case "object":
		if node.Kind != KindMap {
			return nil, fmt.Errorf("expected map, got %s", node.Kind)
		}
		m, ok := node.Value.(*YispMap)
		if !ok {
			return nil, fmt.Errorf("expected map, got %T", node.Value)
		}
		newMap := NewYispMap()
		for key, value := range m.AllFromFront() {
			itemNode, ok := value.(*YispNode)
			if !ok {
				return nil, fmt.Errorf("expected YispNode, got %T", value)
			}
			subSchema, ok := s.Properties[key]
			if !ok {
				if !s.AdditionalProperties {
					return nil, fmt.Errorf("unexpected property: %s", key)
				}
				// If additional properties are allowed, just add the item as is
				newMap.Set(key, itemNode)
				continue
			}
			castedNode, err := subSchema.Cast(itemNode)
			if err != nil {
				return nil, fmt.Errorf("error casting property '%s': %v", key, err)
			}
			newMap.Set(key, castedNode)
		}
		for key, subSchema := range s.Properties {
			if _, ok := m.Get(key); !ok {
				if slices.Contains(s.Required, key) {
					return nil, fmt.Errorf("missing required property: %s", key)
				}
				if subSchema.Default != nil {
					defaultNode := &YispNode{
						Kind:  schemaTypeToKind[subSchema.Type],
						Value: subSchema.Default,
						Tag:   node.Tag,
						Pos:   node.Pos,
						Type:  subSchema,
					}
					newMap.Set(key, defaultNode)
				}
			}
		}
		return &YispNode{
			Kind:  KindMap,
			Value: newMap,
			Tag:   node.Tag,
			Pos:   node.Pos,
			Type:  s,
		}, nil

	case "function":
		return nil, fmt.Errorf("currently, function casting is not supported")
	default:
		// Unknown type, no casting
		return nil, fmt.Errorf("unknown schema type: %s", s.Type)
	}
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
			if !arg.Equals(other.Arguments[i]) {
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
