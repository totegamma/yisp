package yisp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

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
	case map[string]any:
		return len(v) != 0, nil
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

func DeepMerge(dst, src map[string]any) map[string]any {
	for key, value := range src {
		if dstValue, ok := dst[key]; ok {
			if dstMap, ok := dstValue.(map[string]any); ok {
				if srcMap, ok := value.(map[string]any); ok {
					dst[key] = DeepMerge(dstMap, srcMap)
					continue
				}
			}
		}
		dst[key] = value
	}
	return dst
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
