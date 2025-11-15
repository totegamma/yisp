package core

import (
	"testing"
)

func TestValidateWithAllowPartial(t *testing.T) {
	// Create a schema with required properties
	schema := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"name": {
				Type: "string",
			},
			"age": {
				Type: "integer",
			},
		},
		Required: []string{"name", "age"},
	}

	// Create a node with only one of the required properties
	partialMap := NewYispMap()
	partialMap.Set("name", &YispNode{
		Kind:  KindString,
		Value: "John",
	})

	partialNode := &YispNode{
		Kind:  KindMap,
		Value: partialMap,
	}

	// Test 1: Validate with allowPartial=false should fail
	err := schema.ValidateWithOptions(partialNode, false)
	if err == nil {
		t.Error("Expected validation to fail when allowPartial=false with missing required property")
	}

	// Test 2: Validate with allowPartial=true should succeed
	err = schema.ValidateWithOptions(partialNode, true)
	if err != nil {
		t.Errorf("Expected validation to succeed when allowPartial=true, got error: %v", err)
	}

	// Test 3: Validate with all required properties should succeed with allowPartial=false
	completeMap := NewYispMap()
	completeMap.Set("name", &YispNode{
		Kind:  KindString,
		Value: "Jane",
	})
	completeMap.Set("age", &YispNode{
		Kind:  KindInt,
		Value: 30,
	})

	completeNode := &YispNode{
		Kind:  KindMap,
		Value: completeMap,
	}

	err = schema.ValidateWithOptions(completeNode, false)
	if err != nil {
		t.Errorf("Expected validation to succeed with all required properties, got error: %v", err)
	}

	// Test 4: Original Validate method should behave as allowPartial=false
	err = schema.Validate(partialNode)
	if err == nil {
		t.Error("Expected original Validate method to fail with missing required property")
	}
}

func TestValidateWithAllowPartialNestedObjects(t *testing.T) {
	// Create a schema with nested objects and required properties
	schema := &Schema{
		Type: "object",
		Properties: map[string]*Schema{
			"user": {
				Type: "object",
				Properties: map[string]*Schema{
					"name": {
						Type: "string",
					},
					"email": {
						Type: "string",
					},
				},
				Required: []string{"name", "email"},
			},
		},
		Required: []string{"user"},
	}

	// Create a node with partial nested object
	nestedMap := NewYispMap()
	nestedMap.Set("name", &YispNode{
		Kind:  KindString,
		Value: "Alice",
	})

	rootMap := NewYispMap()
	rootMap.Set("user", &YispNode{
		Kind:  KindMap,
		Value: nestedMap,
	})

	node := &YispNode{
		Kind:  KindMap,
		Value: rootMap,
	}

	// Test with allowPartial=false should fail due to missing nested required property
	err := schema.ValidateWithOptions(node, false)
	if err == nil {
		t.Error("Expected validation to fail with missing nested required property when allowPartial=false")
	}

	// Test with allowPartial=true should succeed
	err = schema.ValidateWithOptions(node, true)
	if err != nil {
		t.Errorf("Expected validation to succeed with allowPartial=true for nested objects, got error: %v", err)
	}
}

func TestValidateWithAllowPartialArrays(t *testing.T) {
	// Create a schema with array items that have required properties
	schema := &Schema{
		Type: "array",
		Items: &Schema{
			Type: "object",
			Properties: map[string]*Schema{
				"id": {
					Type: "integer",
				},
				"value": {
					Type: "string",
				},
			},
			Required: []string{"id", "value"},
		},
	}

	// Create partial array items
	partialItem := NewYispMap()
	partialItem.Set("id", &YispNode{
		Kind:  KindInt,
		Value: 1,
	})

	arr := []any{
		&YispNode{
			Kind:  KindMap,
			Value: partialItem,
		},
	}

	arrayNode := &YispNode{
		Kind:  KindArray,
		Value: arr,
	}

	// Test with allowPartial=false should fail
	err := schema.ValidateWithOptions(arrayNode, false)
	if err == nil {
		t.Error("Expected validation to fail for array items with missing required properties when allowPartial=false")
	}

	// Test with allowPartial=true should succeed
	err = schema.ValidateWithOptions(arrayNode, true)
	if err != nil {
		t.Errorf("Expected validation to succeed for array items with allowPartial=true, got error: %v", err)
	}
}
