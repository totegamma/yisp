package engine

import (
	"encoding/json"
	"fmt"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/internal/yaml"
)

// YamlPrint prints an object as YAML
func YamlPrint(obj any) {
	b, _ := yaml.Marshal(obj)
	fmt.Println(string(b))
}

func PrintYispNode(tag string, node *core.YispNode) {
	native, err := ToNative(node)
	if err != nil {
		fmt.Println("Error converting to native:", err)
		return
	}
	b, _ := json.MarshalIndent(native, "", "  ")
	fmt.Println(tag, string(b))
}

func EvalAndCastNode[T any](node *core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (T, error) {
	evaluated, err := e.Eval(node, env, mode)
	if err != nil {
		return *new(T), err
	}

	castedValue, ok := evaluated.Value.(T)
	if !ok {
		return *new(T), fmt.Errorf("expected value of type %T but got %T", new(T), evaluated.Value)
	}

	return castedValue, nil
}

func EvalAndCastAny[T any](value any, env *core.Env, mode core.EvalMode, e core.Engine) (T, error) {

	node, ok := value.(*core.YispNode)
	if !ok {
		return *new(T), fmt.Errorf("expected core.YispNode but got %T", value)
	}

	return EvalAndCastNode[T](node, env, mode, e)
}

// compareValues compares two values of any type for equality
// It only compares values of the same type
func compareValues(cdr []*core.YispNode, opName string, expectEqual bool) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("%s requires 2 arguments, got %d", opName, len(cdr)))
	}

	firstNode := cdr[0]
	secondNode := cdr[1]

	// Only compare values of the same type
	equal := false

	// Handle different type combinations
	switch v1 := firstNode.Value.(type) {
	case int:
		switch v2 := secondNode.Value.(type) {
		case int:
			equal = v1 == v2
		case float64:
			equal = float64(v1) == v2
		default:
			equal = false // Different types
		}
	case float64:
		switch v2 := secondNode.Value.(type) {
		case int:
			equal = v1 == float64(v2)
		case float64:
			equal = v1 == v2
		default:
			equal = false // Different types
		}
	case string:
		switch v2 := secondNode.Value.(type) {
		case string:
			equal = v1 == v2
		default:
			equal = false // Different types
		}
	case bool:
		switch v2 := secondNode.Value.(type) {
		case bool:
			equal = v1 == v2
		default:
			equal = false // Different types
		}
	default:
		// For other types, we just check if they're the same type and value
		equal = firstNode.Value == secondNode.Value
	}

	// For != operation, invert the result
	if !expectEqual {
		equal = !equal
	}

	return &core.YispNode{
		Kind:  core.KindBool,
		Value: equal,
	}, nil
}

// compareNumbers compares two numbers using the provided comparison function
// It handles both integers and floating point numbers
func compareNumbers(cdr []*core.YispNode, opName string, cmp func(float64, float64) bool) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("%s requires 2 arguments, got %d", opName, len(cdr)))
	}

	firstNode := cdr[0]
	var firstNum float64
	switch v := firstNode.Value.(type) {
	case int:
		firstNum = float64(v)
	case float64:
		firstNum = v
	default:
		return nil, core.NewEvaluationError(firstNode, fmt.Sprintf("invalid first argument type for %s: %T (value: %v)", opName, firstNode.Value, firstNode.Value))
	}

	secondNode := cdr[1]
	var secondNum float64
	switch v := secondNode.Value.(type) {
	case int:
		secondNum = float64(v)
	case float64:
		secondNum = v
	default:
		return nil, core.NewEvaluationError(secondNode, fmt.Sprintf("invalid second argument type for %s: %T (value: %v)", opName, secondNode.Value, secondNode.Value))
	}

	return &core.YispNode{
		Kind:  core.KindBool,
		Value: cmp(firstNum, secondNum),
	}, nil
}

// isTruthy determines if a value is considered "truthy" in a boolean context
func isTruthy(node *core.YispNode) (bool, error) {
	switch node.Kind {
	case core.KindNull:
		return false, nil
	case core.KindBool:
		v, ok := node.Value.(bool)
		if !ok {
			return false, fmt.Errorf("expected bool, got %T", node.Value)
		}
		return v, nil
	case core.KindInt:
		v, ok := node.Value.(int)
		if !ok {
			return false, fmt.Errorf("expected int, got %T", node.Value)
		}
		return v != 0, nil
	case core.KindFloat:
		v, ok := node.Value.(float64)
		if !ok {
			return false, fmt.Errorf("expected float64, got %T", node.Value)
		}
		return v != 0.0, nil
	case core.KindString:
		v, ok := node.Value.(string)
		if !ok {
			return false, fmt.Errorf("expected string, got %T", node.Value)
		}
		return v != "", nil
	case core.KindArray:
		v, ok := node.Value.([]any)
		if !ok {
			return false, fmt.Errorf("expected []any, got %T", node.Value)
		}
		return len(v) != 0, nil
	case core.KindMap:
		v, ok := node.Value.(*core.YispMap)
		if !ok {
			return false, fmt.Errorf("expected *core.YispMap, got %T", node.Value)
		}
		return v.Len() != 0, nil
	case core.KindLambda:
		// Lambda functions are always considered isTruthy
		return true, nil
	case core.KindParameter:
		// Parameters are always considered isTruthy
		return true, nil
	case core.KindSymbol:
		// Symbols are always considered isTruthy
		return true, nil
	case core.KindType:
		// Types are always considered isTruthy
		return true, nil
	default:
		// Any other non-nil value is considered isTruthy
		if node.Value != nil {
			return true, nil
		}
		return false, nil
	}
}

func pad(length int) string {
	result := ""
	for range length {
		result += "  "
	}
	return result
}

func ToNative(node *core.YispNode) (any, error) {
	switch node.Kind {
	case core.KindNull, core.KindBool, core.KindInt, core.KindFloat, core.KindString:
		return node.Value, nil
	case core.KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array value. Actual type: %T", node.Value)
		}
		results := make([]any, len(arr))
		for i, item := range arr {
			node, ok := item.(*core.YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}
			var err error
			results[i], err = ToNative(node)
			if err != nil {
				return nil, err
			}
		}
		return results, nil
	case core.KindMap:
		m, ok := node.Value.(*core.YispMap)
		if !ok {
			return nil, fmt.Errorf("invalid map value")
		}
		results := map[string]any{}
		for key, item := range m.AllFromFront() {
			node, ok := item.(*core.YispNode)
			if !ok {
				return nil, fmt.Errorf("invalid item type: %T", item)
			}

			content, err := ToNative(node)
			if err != nil {
				return nil, err
			}

			results[key] = content

		}
		return results, nil

	case core.KindLambda:
		return "(lambda)", nil
	case core.KindParameter:
		return "(parameter)", nil
	case core.KindSymbol:
		return "*" + node.Value.(string), nil
	case core.KindType:
		return "(type)", nil
	default:
		return "(unknown)", nil
	}
}
