package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	ID                   string             `json:"$id,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AdditionalProperties any                `json:"additionalProperties,omitempty"`
	Arguments            []*Schema          `json:"arguments,omitempty"`
	Returns              *Schema            `json:"returns,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              any                `json:"default,omitempty"`

	OneOf []*Schema `json:"oneOf,omitempty"`

	// Numeric constraints
	MultipleOf       *int     `json:"multipleOf,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`

	// String constraints
	MinLength *int `json:"minLength,omitempty"`
	MaxLength *int `json:"maxLength,omitempty"`

	PatchStrategy    string `json:"patchStrategy,omitempty"`
	PatchMergeKey    string `json:"patchMergeKey,omitempty"`
	K8sPatchStrategy string `json:"x-kubernetes-patch-strategy,omitempty"`
	K8sPatchMergeKey string `json:"x-kubernetes-patch-merge-key,omitempty"`
}

func (s *Schema) GetProperties() map[string]*Schema {
	if s.Properties == nil {
		return make(map[string]*Schema)
	}

	newProperties := make(map[string]*Schema, len(s.Properties))
	for key, value := range s.Properties {
		if value.Ref != "" {
			refSchema, err := LoadSchemaFromID(value.Ref)
			if err != nil {
				fmt.Printf("Error loading schema from ref %s: %v\n", value.Ref, err)
				continue
			}
			newProperties[key] = refSchema
		} else {
			newProperties[key] = value
		}
	}

	return newProperties
}

func (s *Schema) GetItems() *Schema {
	if s.Items != nil && s.Items.Ref != "" {
		refSchema, err := LoadSchemaFromID(s.Items.Ref)
		if err != nil {
			fmt.Printf("Error loading schema from ref %s: %v\n", s.Items.Ref, err)
			return nil
		}
		return refSchema
	}
	return s.Items
}

func (s *Schema) GetAdditionalProperties() any {

	if s.AdditionalProperties == nil {
		return false
	}

	switch ap := s.AdditionalProperties.(type) {
	case bool:
		return ap
	default:
		jsonStr, err := json.Marshal(s.AdditionalProperties)
		if err != nil {
			return fmt.Sprintf("failed to marshal additionalProperties: %v", err)
		}
		var additionalSchema *Schema
		err = json.Unmarshal(jsonStr, &additionalSchema)
		if err != nil {
			return fmt.Sprintf("failed to unmarshal additionalProperties: %v", err)
		}

		if additionalSchema.Ref != "" {
			refSchema, err := LoadSchemaFromID(additionalSchema.Ref)
			if err != nil {
				fmt.Printf("Error loading schema from ref %s: %v\n", additionalSchema.Ref, err)
			}
			return refSchema
		}

		return additionalSchema
	}
}

func (s *Schema) GetPatchStrategy() string {
	if s.PatchStrategy != "" {
		return s.PatchStrategy
	}
	if s.K8sPatchStrategy != "" {
		return s.K8sPatchStrategy
	}
	return ""
}

func (s *Schema) GetPatchMergeKey() string {
	if s.PatchMergeKey != "" {
		return s.PatchMergeKey
	}
	if s.K8sPatchMergeKey != "" {
		return s.K8sPatchMergeKey
	}
	return ""
}

func (s *Schema) ToYispNode() (*YispNode, error) {
	jsonStr, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var anyValue any
	err = json.Unmarshal(jsonStr, &anyValue)
	if err != nil {
		return nil, err
	}

	return ParseAny("", anyValue)
}

func LoadSchemaFromURL(url string) (*Schema, error) {

	cachekey := base64.StdEncoding.EncodeToString([]byte(url))
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	schemasPath := filepath.Join(home, ".cache", "yisp", "schemas", cachekey+".json")
	file, err := os.Open(schemasPath)
	if err == nil {
		defer file.Close()
		var schema Schema
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&schema)
		if err != nil {
			return nil, err
		}
		return &schema, nil
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch schema from URL: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var schema Schema
		err = json.Unmarshal(body, &schema)
		if err != nil {
			return nil, err
		}

		err = os.MkdirAll(filepath.Dir(schemasPath), os.ModePerm)
		if err != nil {
			return nil, err
		}

		file, err := os.Create(schemasPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		err = encoder.Encode(&schema)
		if err != nil {
			return nil, err
		}

		return &schema, nil
	}
}

func LoadSchemaFromID(id string) (*Schema, error) {

	if strings.HasPrefix(id, "#") {
		split := strings.Split(id, "/")
		id = split[len(split)-1]
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	schemasPath := filepath.Join(home, ".cache", "yisp", "schemas", id+".json")
	file, err := os.Open(schemasPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var schema Schema
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func LoadSchemaFromGVK(group, version, kind string) (*Schema, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	gvkPath := filepath.Join(home, ".cache", "yisp", "gvk", fmt.Sprintf("%s_%s_%s.txt", group, version, kind))
	file, err := os.Open(gvkPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	id, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return LoadSchemaFromID(string(id))
}

func (s *Schema) Validate(node *YispNode) error {
	return s.ValidateWithOptions(node, false)
}

func (s *Schema) ValidateWithOptions(node *YispNode, allowPartial bool) error {

	if s.OneOf != nil {
		var errors []string
		for _, subSchema := range s.OneOf {
			err := subSchema.ValidateWithOptions(node, allowPartial)
			if err == nil {
				return nil // Valid against one of the schemas
			}
			errors = append(errors, err.Error())
		}
		return NewEvaluationError(node, fmt.Sprintf("node does not match any of the oneOf schemas: %s", strings.Join(errors, ", ")))
	}

	switch s.Type {
	case "any":
		return nil
	case "null":
		if node.Kind != KindNull {
			return NewEvaluationError(node, fmt.Sprintf("expected null, got %s", node.Kind))
		}
	case "boolean":
		if node.Kind != KindBool {
			return NewEvaluationError(node, fmt.Sprintf("expected bool, got %s", node.Kind))
		}
	case "integer":
		if node.Kind != KindInt {
			return NewEvaluationError(node, fmt.Sprintf("expected int, got %s", node.Kind))
		}
		if s.Minimum != nil && node.Value.(int) < int(*s.Minimum) {
			return NewEvaluationError(node, fmt.Sprintf("value %d is less than minimum %f", node.Value.(int), *s.Minimum))
		}
		if s.Maximum != nil && node.Value.(int) > int(*s.Maximum) {
			return NewEvaluationError(node, fmt.Sprintf("value %d is greater than maximum %f", node.Value.(int), *s.Maximum))
		}
		if s.ExclusiveMinimum != nil && node.Value.(int) <= int(*s.ExclusiveMinimum) {
			return NewEvaluationError(node, fmt.Sprintf("value %d is not greater than exclusive minimum %f", node.Value.(int), *s.ExclusiveMinimum))
		}
		if s.ExclusiveMaximum != nil && node.Value.(int) >= int(*s.ExclusiveMaximum) {
			return NewEvaluationError(node, fmt.Sprintf("value %d is not less than exclusive maximum %f", node.Value.(int), *s.ExclusiveMaximum))
		}
		if s.MultipleOf != nil {
			if node.Value.(int)%*s.MultipleOf != 0 {
				return NewEvaluationError(node, fmt.Sprintf("value %d is not a multiple of %d", node.Value.(int), *s.MultipleOf))
			}
		}
	case "float":
		if node.Kind != KindFloat {
			return NewEvaluationError(node, fmt.Sprintf("expected float, got %s", node.Kind))
		}
		if s.Minimum != nil && node.Value.(float64) < *s.Minimum {
			return NewEvaluationError(node, fmt.Sprintf("value %f is less than minimum %f", node.Value.(float64), *s.Minimum))
		}
		if s.Maximum != nil && node.Value.(float64) > *s.Maximum {
			return NewEvaluationError(node, fmt.Sprintf("value %f is greater than maximum %f", node.Value.(float64), *s.Maximum))
		}
		if s.ExclusiveMinimum != nil && node.Value.(float64) <= *s.ExclusiveMinimum {
			return NewEvaluationError(node, fmt.Sprintf("value %f is not greater than exclusive minimum %f", node.Value.(float64), *s.ExclusiveMinimum))
		}
		if s.ExclusiveMaximum != nil && node.Value.(float64) >= *s.ExclusiveMaximum {
			return NewEvaluationError(node, fmt.Sprintf("value %f is not less than exclusive maximum %f", node.Value.(float64), *s.ExclusiveMaximum))
		}
		if s.MultipleOf != nil {
			if int(node.Value.(float64))%*s.MultipleOf != 0 {
				return NewEvaluationError(node, fmt.Sprintf("value %f is not a multiple of %d", node.Value.(float64), *s.MultipleOf))
			}
		}
	case "string":
		if node.Kind != KindString {
			if s.Format == "int-or-string" {
				if node.Kind != KindInt {
					return NewEvaluationError(node, fmt.Sprintf("expected string or int, got %s", node.Kind))
				}
				node.Kind = KindString
				node.Value = fmt.Sprintf("%d", node.Value.(int))
			} else {
				return NewEvaluationError(node, fmt.Sprintf("expected string, got %s", node.Kind))
			}
		}
		if s.MinLength != nil && len(node.Value.(string)) < *s.MinLength {
			return NewEvaluationError(node, fmt.Sprintf("string length %d is less than minimum %d", len(node.Value.(string)), *s.MinLength))
		}
		if s.MaxLength != nil && len(node.Value.(string)) > *s.MaxLength {
			return NewEvaluationError(node, fmt.Sprintf("string length %d is greater than maximum %d", len(node.Value.(string)), *s.MaxLength))
		}
	case "array":
		if node.Kind != KindArray {
			return NewEvaluationError(node, fmt.Sprintf("expected array, got %s", node.Kind))
		}
		if s.Items != nil {
			subSchema := s.GetItems()
			arr, ok := node.Value.([]any)
			if !ok {
				return NewEvaluationError(node, fmt.Sprintf("expected array, got %T", node.Value))
			}

			for _, item := range arr {
				itemNode, ok := item.(*YispNode)
				if !ok {
					return NewEvaluationError(node, fmt.Sprintf("expected YispNode, got %T", item))
				}
				if err := subSchema.ValidateWithOptions(itemNode, allowPartial); err != nil {
					return err
				}
			}
		}
	case "object":
		if node.Kind != KindMap {
			return NewEvaluationError(node, fmt.Sprintf("expected map, got %s", node.Kind))
		}
		m, ok := node.Value.(*YispMap)
		if !ok {
			return NewEvaluationError(node, fmt.Sprintf("expected map, got %T", node.Value))
		}

		processed := make(map[string]bool)
		for key, subSchema := range s.GetProperties() {
			item, ok := m.Get(key)
			if !ok {
				if slices.Contains(s.Required, key) && !allowPartial {
					return NewEvaluationError(node, fmt.Sprintf("missing required property: %s", key))
				} else {
					continue
				}
			}
			itemNode, ok := item.(*YispNode)
			if !ok {
				return NewEvaluationError(node, fmt.Sprintf("[object]expected YispNode, got %T", item))
			}

			if err := subSchema.ValidateWithOptions(itemNode, allowPartial); err != nil {
				return err
			}
			processed[key] = true
		}

		left := NewYispMap()
		for key, item := range m.AllFromFront() {
			if _, ok := processed[key]; !ok {
				if key != "$schema" {
					left.Set(key, item)
				}
			}
		}

		if left.Len() != 0 {
			additionalProperties := s.GetAdditionalProperties()

			switch ap := additionalProperties.(type) {
			case bool:
				if !ap {
					return NewEvaluationError(node, fmt.Sprintf("unexpected properties: %v", left.Keys()))
				}
			case *Schema:
				for key, item := range left.AllFromFront() {
					itemNode, ok := item.(*YispNode)
					if !ok {
						return NewEvaluationError(node, fmt.Sprintf("expected YispNode, got %T", item))
					}
					if err := ap.ValidateWithOptions(itemNode, allowPartial); err != nil {
						return NewEvaluationError(node, fmt.Sprintf("additional property %s does not match schema: %v", key, err))
					}
				}
			default:
				return NewEvaluationError(node, fmt.Sprintf("unexpected additionalProperties type: %T", ap))
			}
		}

	case "function":
		if node.Kind != KindLambda {
			return NewEvaluationError(node, fmt.Sprintf("expected function, got %s", node.Kind))
		}
		fn, ok := node.Value.(*Lambda)
		if !ok {
			return NewEvaluationError(node, fmt.Sprintf("expected YispLambda, got %T", node.Value))
		}
		if len(fn.Arguments) != len(s.Arguments) {
			return NewEvaluationError(node, fmt.Sprintf("expected %d arguments, got %d", len(s.Arguments), len(fn.Arguments)))
		}
		for i, arg := range s.Arguments {
			if fn.Arguments[i].Schema != nil && !arg.Equals(fn.Arguments[i].Schema) {
				return NewEvaluationError(node, fmt.Sprintf("argument %d does not match schema", i))
			}

		}
		if s.Returns != nil && fn.Returns != nil {
			if !s.Returns.Equals(fn.Returns) {
				return NewEvaluationError(node, fmt.Sprintf("return type does not match schema. Expected %s, got %s", s.Returns.Type, fn.Returns.Type))
			}
		}

	default:
		JsonPrint("schema", s)
		return NewEvaluationError(node, fmt.Sprintf("unknown type: %s", s.Type))
	}

	return nil
}

func (s *Schema) InterpolateDefaults(node *YispNode) error {
	switch s.Type {
	case "array":
		if node.Kind != KindArray {
			return NewEvaluationError(node, fmt.Sprintf("expected array, got %s", node.Kind))
		}
		if s.Items == nil {
			return nil // No items schema to interpolate defaults for
		}
		arr, ok := node.Value.([]any)
		if !ok {
			return NewEvaluationError(node, fmt.Sprintf("expected array, got %T", node.Value))
		}
		for i, item := range arr {
			itemNode, ok := item.(*YispNode)
			if !ok {
				return NewEvaluationError(node, fmt.Sprintf("expected YispNode, got %T", item))
			}
			if err := s.Items.InterpolateDefaults(itemNode); err != nil {
				return NewEvaluationError(node, fmt.Sprintf("failed to interpolate defaults for array item %d: %v", i, err))
			}
		}
		return nil
	case "object":
		if node.Kind != KindMap {
			return NewEvaluationError(node, fmt.Sprintf("expected map, got %s", node.Kind))
		}
		if node.Value == nil {
			node.Value = NewYispMap()
		}
		m, ok := node.Value.(*YispMap)
		if !ok {
			return NewEvaluationError(node, fmt.Sprintf("expected map, got %T", node.Value))
		}
		for key, subSchema := range s.Properties {
			item, ok := m.Get(key)
			if !ok {
				if slices.Contains(s.Required, key) {
					return NewEvaluationError(node, fmt.Sprintf("missing required property: %s", key))
				}
				if subSchema.Default != nil {
					defaultNode := &YispNode{
						Kind:  schemaTypeToKind[subSchema.Type],
						Value: subSchema.Default,
						Tag:   node.Tag,
						Attr:  node.Attr,
						Type:  subSchema,
					}
					m.Set(key, defaultNode)
					continue
				}
				dummyNode := &YispNode{
					Kind: schemaTypeToKind[subSchema.Type],
				}
				err := subSchema.InterpolateDefaults(dummyNode)
				if err != nil {
					return NewEvaluationError(node, fmt.Sprintf("failed to interpolate dummy Node %v", err))
				}
				if !IsZero(dummyNode.Value) {
					m.Set(key, dummyNode)
				}
				continue
			}
			itemNode, ok := item.(*YispNode)
			if !ok {
				return NewEvaluationError(node, fmt.Sprintf("expected YispNode, got %T", item))
			}
			if err := subSchema.InterpolateDefaults(itemNode); err != nil {
				return NewEvaluationError(node, fmt.Sprintf("failed to interpolate defaults for property %s: %v", key, err))
			}
		}
		return nil
	default:
		return nil
	}
}

func (s *Schema) Cast(node *YispNode) (*YispNode, error) {
	err := s.InterpolateDefaults(node)
	if err != nil {
		return nil, fmt.Errorf("failed to cast %v to %v (%v)", node, s, err)
	}
	err = s.Validate(node)
	if err != nil {
		return nil, fmt.Errorf("failed to cast %v to %v (%v)", node, s, err)
	}

	node.Type = s

	return node, nil

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
