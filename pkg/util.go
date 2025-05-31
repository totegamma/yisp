package yisp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/totegamma/yisp/internal/k8stypes"
	"github.com/totegamma/yisp/yaml"
)

// JsonPrint prints an object as formatted JSON with a tag
func JsonPrint(tag string, obj any) {
	b, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Println(tag, string(b))
}

// YamlPrint prints an object as YAML
func YamlPrint(obj any) {
	b, _ := yaml.Marshal(obj)
	fmt.Println(string(b))
}

func PrintYispNode(tag string, node *YispNode) {
	native, err := ToNative(node)
	if err != nil {
		fmt.Println("Error converting to native:", err)
		return
	}
	b, _ := json.MarshalIndent(native, "", "  ")
	fmt.Println(tag, string(b))
}

func EvalAndCastNode[T any](node *YispNode, env *Env, mode EvalMode) (T, error) {
	evaluated, err := Eval(node, env, mode)
	if err != nil {
		return *new(T), err
	}

	castedValue, ok := evaluated.Value.(T)
	if !ok {
		return *new(T), fmt.Errorf("expected value of type %T but got %T", new(T), evaluated.Value)
	}

	return castedValue, nil
}

func EvalAndCastAny[T any](value any, env *Env, mode EvalMode) (T, error) {

	node, ok := value.(*YispNode)
	if !ok {
		return *new(T), fmt.Errorf("expected YispNode but got %T", value)
	}

	return EvalAndCastNode[T](node, env, mode)
}

// compareValues compares two values of any type for equality
// It only compares values of the same type
func compareValues(cdr []*YispNode, env *Env, mode EvalMode, opName string, expectEqual bool) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("%s requires 2 arguments, got %d", opName, len(cdr)))
	}

	firstNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate first argument"), err)
	}

	secondNode, err := Eval(cdr[1], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[1], fmt.Sprintf("failed to evaluate second argument"), err)
	}

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

	return &YispNode{
		Kind:  KindBool,
		Value: equal,
	}, nil
}

// compareNumbers compares two numbers using the provided comparison function
// It handles both integers and floating point numbers
func compareNumbers(cdr []*YispNode, env *Env, mode EvalMode, opName string, cmp func(float64, float64) bool) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("%s requires 2 arguments, got %d", opName, len(cdr)))
	}

	firstNode, err := Eval(cdr[0], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[0], fmt.Sprintf("failed to evaluate first argument"), err)
	}

	var firstNum float64
	switch v := firstNode.Value.(type) {
	case int:
		firstNum = float64(v)
	case float64:
		firstNum = v
	default:
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid first argument type for %s: %T (value: %v)", opName, firstNode.Value, firstNode.Value))
	}

	secondNode, err := Eval(cdr[1], env, mode)
	if err != nil {
		return nil, NewEvaluationErrorWithParent(cdr[1], fmt.Sprintf("failed to evaluate second argument"), err)
	}

	var secondNum float64
	switch v := secondNode.Value.(type) {
	case int:
		secondNum = float64(v)
	case float64:
		secondNum = v
	default:
		return nil, NewEvaluationError(secondNode, fmt.Sprintf("invalid second argument type for %s: %T (value: %v)", opName, secondNode.Value, secondNode.Value))
	}

	return &YispNode{
		Kind:  KindBool,
		Value: cmp(firstNum, secondNum),
	}, nil
}

// isTruthy determines if a value is considered "truthy" in a boolean context
func isTruthy(node *YispNode) (bool, error) {
	switch v := node.Value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case float64:
		return v != 0.0, nil
	case string:
		// Any non-empty string is truthy
		return v != "", nil
	case []any:
		return len(v) != 0, nil
	case *YispMap:
		return v.Len() != 0, nil
	case *Lambda:
		// Lambda functions are always truthy
		return true, nil
	case nil:
		return false, nil
	default:
		// Any other non-nil value is considered truthy
		return true, nil
	}
}

func DeepMergeYispNode(dst, src *YispNode) (*YispNode, error) {
	if dst.Kind == KindMap && src.Kind == KindMap {

		dstMap, dstOK := dst.Value.(*YispMap)
		srcMap, srcOK := src.Value.(*YispMap)
		if !dstOK || !srcOK {
			return nil, fmt.Errorf("invalid map value. Actual type: %T", dst.Value)
		}

		allKeys := make([]string, 0)
		for key := range dstMap.Keys() {
			if !slices.Contains(allKeys, key) {
				allKeys = append(allKeys, key)
			}
		}
		for key := range srcMap.Keys() {
			if !slices.Contains(allKeys, key) {
				allKeys = append(allKeys, key)
			}
		}
		result := NewYispMap()
		for _, key := range allKeys {
			dstVal, dstOK := dstMap.Get(key)
			srcVal, srcOK := srcMap.Get(key)

			if dstOK && srcOK {
				dstNode, dstNodeOK := dstVal.(*YispNode)
				srcNode, srcNodeOK := srcVal.(*YispNode)

				if dstNodeOK && srcNodeOK {
					mergedNode, err := DeepMergeYispNode(dstNode, srcNode)
					if err != nil {
						return nil, err
					}
					result.Set(key, mergedNode)
				}
			} else if dstOK {
				result.Set(key, dstVal)
			} else if srcOK {
				result.Set(key, srcVal)
			}
		}

		return &YispNode{
			Kind:  KindMap,
			Value: result,
		}, nil

	} else if dst.Kind == KindArray && src.Kind == KindArray {

		dstArray, dstOK := dst.Value.([]any)
		srcArray, srcOK := src.Value.([]any)
		if !dstOK || !srcOK {
			return nil, fmt.Errorf("invalid array value. Actual type: %T", dst.Value)
		}
		result := make([]any, len(dstArray)+len(srcArray))
		copy(result, dstArray)
		copy(result[len(dstArray):], srcArray)
		return &YispNode{
			Kind:  KindArray,
			Value: result,
		}, nil

	} else {
		return src, nil
	}
}

func RenderCode(file string, line, after, before int, comments []Comment) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	startLine := max(line-before, 1)

	scanner := bufio.NewScanner(f)
	for range startLine - 1 {
		if !scanner.Scan() {
			break
		}
	}
	result := ""

	result += file + "\n"
	for range len(file) {
		result += "="
	}
	result += "\n"

	lnFormat := "%d |"
	maxLineNumberLen := len(fmt.Sprintf(lnFormat, line+after))

	for i := range after + before + 1 {
		if !scanner.Scan() {
			break
		}

		currentLine := startLine + i

		ln := fmt.Sprintf(lnFormat, currentLine)
		for range maxLineNumberLen - len(ln) {
			ln = " " + ln
		}

		result += fmt.Sprintf("%s%s\n", ln, scanner.Text())
		for _, comment := range comments {
			if comment.Line == currentLine {
				for range comment.Column - 1 + len(ln) {
					result += " "
				}
				result += "\x1b[31m^ " + comment.Text + "\x1b[0m\n"
			}
		}

	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result, nil
}

func pad(length int) string {
	result := ""
	for range length {
		result += "  "
	}
	return result
}

func ToNative(node *YispNode) (any, error) {
	switch node.Kind {
	case KindNull, KindBool, KindInt, KindFloat, KindString:
		return node.Value, nil
	case KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid array value. Actual type: %T", node.Value)
		}
		results := make([]any, len(arr))
		for i, item := range arr {
			node, ok := item.(*YispNode)
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
	case KindMap:
		m, ok := node.Value.(*YispMap)
		if !ok {
			return nil, fmt.Errorf("invalid map value")
		}
		results := map[string]any{}
		for key, item := range m.AllFromFront() {
			node, ok := item.(*YispNode)
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

	case KindLambda:
		return "(lambda)", nil
	case KindParameter:
		return "(parameter)", nil
	case KindSymbol:
		return "(symbol)", nil
	case KindType:
		return "(type)", nil
	default:
		return "(unknown)", nil
	}
}

func GetK8sSchema(group, version, kind string) (*Schema, error) {
	schemaBytes, err := k8stypes.GetSchema(group, version, kind)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for %s/%s/%s: %w", group, version, kind, err)
	}
	var schema Schema
	err = json.Unmarshal(schemaBytes, &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema for %s/%s/%s: %w", group, version, kind, err)
	}

	return &schema, nil
}
