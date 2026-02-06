package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
)

// JsonPrint prints an object as formatted JSON with a tag
func JsonPrint(tag string, obj any) {
	b, _ := json.MarshalIndent(obj, "", "  ")
	fmt.Println(tag, string(b))
}

func VerifyTypes(node *YispNode, allowUntypedManifest bool) error {
	// If this node has a schema attached, validate it
	if node.Type != nil && !allowUntypedManifest {
		err := node.Type.Validate(node)
		if err != nil {
			return err
		}
	}

	// Recursively validate child nodes
	switch node.Kind {
	case KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return fmt.Errorf("invalid array value: %T", node.Value)
		}
		for _, item := range arr {
			childNode, ok := item.(*YispNode)
			if !ok {
				continue // Skip non-YispNode items
			}
			if err := VerifyTypes(childNode, allowUntypedManifest); err != nil {
				return err
			}
		}
	case KindMap:
		m, ok := node.Value.(*YispMap)
		if !ok {
			return fmt.Errorf("invalid map value: %T", node.Value)
		}
		for _, item := range m.AllFromFront() {
			childNode, ok := item.(*YispNode)
			if !ok {
				continue // Skip non-YispNode items
			}
			if err := VerifyTypes(childNode, allowUntypedManifest); err != nil {
				return err
			}
		}
	}

	return nil
}

func DeepMergeYispNode(dst, src *YispNode, schema *Schema) (*YispNode, error) {

	strategy := "replace"
	mergeKey := ""
	if schema != nil {
		if schema.GetPatchStrategy() != "" {
			strategy = schema.GetPatchStrategy()
		}
		mergeKey = schema.GetPatchMergeKey()
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

		properties := make(map[string]*Schema)
		if schema != nil {
			properties = schema.GetProperties()
		}

		result := NewYispMap()
		for _, key := range allKeys {
			dstVal, dstOK := dstMap.Get(key)
			srcVal, srcOK := srcMap.Get(key)

			if dstOK && srcOK {
				dstNode, dstNodeOK := dstVal.(*YispNode)
				srcNode, srcNodeOK := srcVal.(*YispNode)

				if dstNodeOK && srcNodeOK {
					mergedNode, err := DeepMergeYispNode(dstNode, srcNode, properties[key])
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
			Type:  src.Type, // TODO: sum type
		}, nil

	} else if dst.Kind == KindArray && src.Kind == KindArray {

		dstArray, dstOK := dst.Value.([]any)
		srcArray, srcOK := src.Value.([]any)
		if !dstOK || !srcOK {
			return nil, fmt.Errorf("invalid array value. Actual type: %T", dst.Value)
		}

		var subSchema *Schema
		if schema != nil {
			subSchema = schema.GetItems()
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
							mergedNode, err := DeepMergeYispNode(dstNode, srcNode, subSchema)
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
			Type:  src.Type, // TODO: sum type
		}, nil

	} else {
		return src, nil
	}
}

func IsZero(v any) bool {
	if v == nil {
		return true
	}

	switch v := v.(type) {
	case *YispMap:
		return v.Len() == 0
	default:
		panic("iszero")
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

func CallEngineByPath(path, base string, env *Env, e Engine) (*YispNode, error) {

	var reader io.Reader

	extension := ".yisp"
	source := ""

	if path == "-" {
		var err error
		reader = os.Stdin
		source, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %v", err)
		}
	} else {
		targetURL, err := url.Parse(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %v", err)
		}

		if base != "" {
			baseURL, err := url.Parse(base)
			if err != nil {
				return nil, fmt.Errorf("failed to parse base URL: %v", err)
			}
			targetURL = baseURL.ResolveReference(targetURL)
		}

		if targetURL.Scheme == "http" || targetURL.Scheme == "https" {
			reader, err = fetchRemote(targetURL.String())
			if err != nil {
				return nil, fmt.Errorf("failed to fetch remote file: %v", err)
			}
		} else {

			stat, err := os.Stat(targetURL.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file: %v", err)
			}

			if stat.IsDir() {
				targetURL = &url.URL{Path: filepath.Join(targetURL.Path, "index.yisp")}
			}
			reader, err = os.Open(targetURL.Path)
			if err != nil {
				return nil, fmt.Errorf("failed to open file: %v", err)
			}
		}
		extension = filepath.Ext(targetURL.Path)
		source = targetURL.String()
	}

	switch extension {
	case ".yisp":
		return e.Run(reader, env, source)
	case ".yml", ".yaml":
		return ParseYaml(source, reader)
	case ".json":
		return ParseJson(source, reader)
	default:
		text, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read txt file: %v", err)
		}

		return &YispNode{
			Kind:  KindString,
			Value: string(text),
		}, nil
	}

}

func fetchRemote(rawURL string) (io.ReadCloser, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote file: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch remote file: %s", resp.Status)
	}
	return resp.Body, nil
}

// isTruthy determines if a value is considered "truthy" in a boolean context
func IsTruthy(node *YispNode) (bool, error) {
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
