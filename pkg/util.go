package yisp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/totegamma/yisp/internal/k8stypes"
	"github.com/totegamma/yisp/internal/yaml"
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
func compareValues(cdr []*YispNode, opName string, expectEqual bool) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("%s requires 2 arguments, got %d", opName, len(cdr)))
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

	return &YispNode{
		Kind:  KindBool,
		Value: equal,
	}, nil
}

// compareNumbers compares two numbers using the provided comparison function
// It handles both integers and floating point numbers
func compareNumbers(cdr []*YispNode, opName string, cmp func(float64, float64) bool) (*YispNode, error) {
	if len(cdr) != 2 {
		return nil, NewEvaluationError(cdr[0], fmt.Sprintf("%s requires 2 arguments, got %d", opName, len(cdr)))
	}

	firstNode := cdr[0]
	var firstNum float64
	switch v := firstNode.Value.(type) {
	case int:
		firstNum = float64(v)
	case float64:
		firstNum = v
	default:
		return nil, NewEvaluationError(firstNode, fmt.Sprintf("invalid first argument type for %s: %T (value: %v)", opName, firstNode.Value, firstNode.Value))
	}

	secondNode := cdr[1]
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
	switch node.Kind {
	case KindNull:
		return false, nil
	case KindBool:
		v, ok := node.Value.(bool)
		if !ok {
			return false, fmt.Errorf("expected bool, got %T", node.Value)
		}
		return v, nil
	case KindInt:
		v, ok := node.Value.(int)
		if !ok {
			return false, fmt.Errorf("expected int, got %T", node.Value)
		}
		return v != 0, nil
	case KindFloat:
		v, ok := node.Value.(float64)
		if !ok {
			return false, fmt.Errorf("expected float64, got %T", node.Value)
		}
		return v != 0.0, nil
	case KindString:
		v, ok := node.Value.(string)
		if !ok {
			return false, fmt.Errorf("expected string, got %T", node.Value)
		}
		return v != "", nil
	case KindArray:
		v, ok := node.Value.([]any)
		if !ok {
			return false, fmt.Errorf("expected []any, got %T", node.Value)
		}
		return len(v) != 0, nil
	case KindMap:
		v, ok := node.Value.(*YispMap)
		if !ok {
			return false, fmt.Errorf("expected *YispMap, got %T", node.Value)
		}
		return v.Len() != 0, nil
	case KindLambda:
		// Lambda functions are always considered isTruthy
		return true, nil
	case KindParameter:
		// Parameters are always considered isTruthy
		return true, nil
	case KindSymbol:
		// Symbols are always considered isTruthy
		return true, nil
	case KindType:
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

func DeepMergeYispNode(dst, src *YispNode, schema *Schema) (*YispNode, error) {

	strategy := "replace"
	mergeKey := ""
	if schema != nil {
		if schema.PatchStrategy != "" {
			strategy = schema.PatchStrategy
		}
		mergeKey = schema.PatchMergeKey
	}

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

			var subType *Schema
			if schema != nil {
				subType = schema.Properties[key]
			}

			if dstOK && srcOK {
				dstNode, dstNodeOK := dstVal.(*YispNode)
				srcNode, srcNodeOK := srcVal.(*YispNode)

				if dstNodeOK && srcNodeOK {
					mergedNode, err := DeepMergeYispNode(dstNode, srcNode, subType)
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

		var subType *Schema
		if schema != nil {
			subType = schema.Items
		}

		var result []any
		if strategy == "replace" {
			result = srcArray
		} else if strategy == "merge" {
			if mergeKey == "" {
				result = append(result, dstArray...)
				result = append(result, srcArray...)
			} else {
				result = dstArray
				for _, srcItem := range srcArray {
					srcNode, ok := srcItem.(*YispNode)
					if !ok {
						return nil, fmt.Errorf("invalid item type in srcArray: %T", srcItem)
					}

					srcMap, ok := srcNode.Value.(*YispMap)
					if !ok {
						return nil, fmt.Errorf("expected YispMap in srcArray, got %T", srcNode.Value)
					}

					keyItem, ok := srcMap.Get(mergeKey)
					if !ok {
						return nil, fmt.Errorf("merge key %s not found in srcMap", mergeKey)
					}

					keyNode, ok := keyItem.(*YispNode)
					if !ok {
						return nil, fmt.Errorf("expected YispNode for merge key, got %T", keyItem)
					}

					key, ok := keyNode.Value.(string)
					if !ok {
						return nil, fmt.Errorf("expected string for merge key, got %T", keyNode.Value)
					}

					// Check if the key already exists in the result
					found := false
					for i, dstItem := range result {
						dstNode, ok := dstItem.(*YispNode)
						if !ok {
							return nil, fmt.Errorf("invalid item type in dstArray: %T", dstItem)
						}

						dstMap, ok := dstNode.Value.(*YispMap)
						if !ok {
							return nil, fmt.Errorf("expected YispMap in dstArray, got %T", dstNode.Value)
						}

						existingKeyItem, ok := dstMap.Get(mergeKey)
						if !ok {
							continue
						}

						existingKeyNode, ok := existingKeyItem.(*YispNode)
						if !ok {
							return nil, fmt.Errorf("expected YispNode for existing merge key, got %T", existingKeyItem)
						}

						existingKey, ok := existingKeyNode.Value.(string)
						if !ok {
							return nil, fmt.Errorf("expected string for existing merge key, got %T", existingKeyNode.Value)
						}

						if existingKey == key {
							// Merge the srcMap into the existing dstMap
							mergedNode, err := DeepMergeYispNode(dstNode, srcNode, subType)
							if err != nil {
								return nil, err
							}
							result[i] = mergedNode
							found = true
							break
						}
					}
					if !found {
						// If not found, add the srcNode to the result
						result = append(result, srcNode)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("unknown patch strategy: %s", strategy)
		}

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
		return "*" + node.Value.(string), nil
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

type GVK struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (gvk *GVK) String() string {
	if gvk.Group == "" {
		return fmt.Sprintf("%s/%s", gvk.Kind, gvk.Version)
	}
	return fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind)
}

func (gvk *GVK) Equal(other *GVK) bool {
	if gvk == nil || other == nil {
		return gvk == other
	}
	return gvk.Group == other.Group && gvk.Version == other.Version && gvk.Kind == other.Kind
}

func GetGVK(node *YispNode) (*GVK, error) {
	if node.Kind != KindMap {
		return nil, fmt.Errorf("expected KindMap for GVK, got %s", node.Kind)
	}

	m, ok := node.Value.(*YispMap)
	if !ok {
		return nil, fmt.Errorf("expected YispMap for GVK, got %T", node.Value)
	}

	apiVersionAny, ok := m.Get("apiVersion")
	if !ok {
		return nil, fmt.Errorf("apiVersion not found in GVK map")
	}
	apiVersionNode, ok := apiVersionAny.(*YispNode)
	if !ok {
		return nil, fmt.Errorf("expected YispNode for apiVersion, got %T", apiVersionAny)
	}
	apiVersion, ok := apiVersionNode.Value.(string)
	if !ok {
		return nil, fmt.Errorf("expected string for apiVersion, got %T", apiVersionNode.Value)
	}

	kindAny, ok := m.Get("kind")
	if !ok {
		return nil, fmt.Errorf("kind not found in GVK map")
	}
	kindNode, ok := kindAny.(*YispNode)
	if !ok {
		return nil, fmt.Errorf("expected YispNode for kind, got %T", kindAny)
	}
	kind, ok := kindNode.Value.(string)
	if !ok {
		return nil, fmt.Errorf("expected string for kind, got %T", kindNode.Value)
	}

	group := ""
	version := ""
	split := strings.Split(apiVersion, "/")

	if len(split) == 2 {
		group = split[0]
		version = split[1]
	} else if len(split) == 1 {
		version = split[0]
	}

	return &GVK{
		Group:   group,
		Version: version,
		Kind:    kind,
	}, nil
}

func IsZero(v any) bool {
	if v == nil {
		return true
	}

	switch v.(type) {
	case *YispMap:
		return v.(*YispMap).Len() == 0
	default:
		panic("iszero")
	}
}
