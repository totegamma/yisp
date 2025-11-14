package lib

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("utils", "op-patch", opOpPatch)
}

// parsePointer parses a JSON Pointer (RFC 6901) path into tokens
func parsePointer(path string) ([]string, error) {
	if path == "" {
		return []string{}, nil
	}
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("JSON Pointer must start with '/', got: %s", path)
	}
	if path == "/" {
		return []string{""}, nil
	}
	
	tokens := strings.Split(path[1:], "/")
	// Unescape special characters per RFC 6901
	for i, token := range tokens {
		token = strings.ReplaceAll(token, "~1", "/")
		token = strings.ReplaceAll(token, "~0", "~")
		tokens[i] = token
	}
	return tokens, nil
}

// getValue retrieves a value from a YispNode using a JSON Pointer path
func getValue(node *core.YispNode, path string) (*core.YispNode, error) {
	tokens, err := parsePointer(path)
	if err != nil {
		return nil, err
	}
	
	current := node
	for _, token := range tokens {
		if current.Kind == core.KindMap {
			m, ok := current.Value.(*core.YispMap)
			if !ok {
				return nil, fmt.Errorf("expected map, got %T", current.Value)
			}
			val, ok := m.Get(token)
			if !ok {
				return nil, fmt.Errorf("key not found: %s", token)
			}
			current, ok = val.(*core.YispNode)
			if !ok {
				return nil, fmt.Errorf("expected YispNode, got %T", val)
			}
		} else if current.Kind == core.KindArray {
			arr, ok := current.Value.([]any)
			if !ok {
				return nil, fmt.Errorf("expected array, got %T", current.Value)
			}
			idx, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", token)
			}
			if idx < 0 || idx >= len(arr) {
				return nil, fmt.Errorf("array index out of bounds: %d", idx)
			}
			current, ok = arr[idx].(*core.YispNode)
			if !ok {
				return nil, fmt.Errorf("expected YispNode, got %T", arr[idx])
			}
		} else {
			return nil, fmt.Errorf("cannot navigate through %s", current.Kind)
		}
	}
	return current, nil
}

// addValue adds a value in a YispNode using a JSON Pointer path (RFC 6902 add operation)
// For arrays, supports appending when index equals array length
// Supports wildcard matching with '*' for map keys
func addValue(node *core.YispNode, path string, value *core.YispNode) error {
	tokens, err := parsePointer(path)
	if err != nil {
		return err
	}
	
	if len(tokens) == 0 {
		return fmt.Errorf("cannot replace root node")
	}
	
	// Check if path contains wildcard
	wildcardIndex := -1
	for i, token := range tokens {
		if token == "*" {
			wildcardIndex = i
			break
		}
	}
	
	// If wildcard found, handle it specially
	if wildcardIndex >= 0 {
		return addValueWithWildcard(node, tokens, wildcardIndex, value)
	}
	
	current := node
	for i := 0; i < len(tokens)-1; i++ {
		token := tokens[i]
		if current.Kind == core.KindMap {
			m, ok := current.Value.(*core.YispMap)
			if !ok {
				return fmt.Errorf("expected map, got %T", current.Value)
			}
			val, ok := m.Get(token)
			if !ok {
				return fmt.Errorf("key not found: %s", token)
			}
			current, ok = val.(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", val)
			}
		} else if current.Kind == core.KindArray {
			arr, ok := current.Value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", current.Value)
			}
			idx, err := strconv.Atoi(token)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", token)
			}
			if idx < 0 || idx >= len(arr) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			current, ok = arr[idx].(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", arr[idx])
			}
		} else {
			return fmt.Errorf("cannot navigate through %s", current.Kind)
		}
	}
	
	lastToken := tokens[len(tokens)-1]
	if current.Kind == core.KindMap {
		m, ok := current.Value.(*core.YispMap)
		if !ok {
			return fmt.Errorf("expected map, got %T", current.Value)
		}
		m.Set(lastToken, value)
	} else if current.Kind == core.KindArray {
		arr, ok := current.Value.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", current.Value)
		}
		
		// RFC 6902: "-" means append to end of array
		if lastToken == "-" {
			current.Value = append(arr, value)
		} else {
			idx, err := strconv.Atoi(lastToken)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", lastToken)
			}
			// RFC 6902: allow appending to array when index equals array length
			if idx < 0 || idx > len(arr) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			if idx == len(arr) {
				// Append to array
				current.Value = append(arr, value)
			} else {
				// Insert at index, shifting elements to the right
				newArr := make([]any, len(arr)+1)
				copy(newArr, arr[:idx])
				newArr[idx] = value
				copy(newArr[idx+1:], arr[idx:])
				current.Value = newArr
			}
		}
	} else {
		return fmt.Errorf("cannot set value in %s", current.Kind)
	}
	
	return nil
}

// addValueWithWildcard handles wildcard matching for add operations
func addValueWithWildcard(node *core.YispNode, tokens []string, wildcardIndex int, value *core.YispNode) error {
	// Navigate to the node just before the wildcard
	current := node
	for i := 0; i < wildcardIndex; i++ {
		token := tokens[i]
		if current.Kind == core.KindMap {
			m, ok := current.Value.(*core.YispMap)
			if !ok {
				return fmt.Errorf("expected map, got %T", current.Value)
			}
			val, ok := m.Get(token)
			if !ok {
				return fmt.Errorf("key not found: %s", token)
			}
			current, ok = val.(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", val)
			}
		} else if current.Kind == core.KindArray {
			arr, ok := current.Value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", current.Value)
			}
			idx, err := strconv.Atoi(token)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", token)
			}
			if idx < 0 || idx >= len(arr) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			current, ok = arr[idx].(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", arr[idx])
			}
		} else {
			return fmt.Errorf("cannot navigate through %s", current.Kind)
		}
	}
	
	// At this point, current is the node where wildcard should match
	if current.Kind == core.KindMap {
		m, ok := current.Value.(*core.YispMap)
		if !ok {
			return fmt.Errorf("expected map, got %T", current.Value)
		}
		
		// If this is the last token, add/merge value to all matching keys
		if wildcardIndex == len(tokens)-1 {
			// Iterate over all keys and add/merge the value
			for key, val := range m.AllFromFront() {
				targetNode, ok := val.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", val)
				}
				
				// If value is a map, merge it with the existing value
				if value.Kind == core.KindMap && targetNode.Kind == core.KindMap {
					targetMap, ok := targetNode.Value.(*core.YispMap)
					if !ok {
						return fmt.Errorf("expected map, got %T", targetNode.Value)
					}
					valueMap, ok := value.Value.(*core.YispMap)
					if !ok {
						return fmt.Errorf("expected map, got %T", value.Value)
					}
					// Merge value map into target map
					for vKey, vVal := range valueMap.AllFromFront() {
						targetMap.Set(vKey, vVal)
					}
				} else {
					// Replace the entire value
					m.Set(key, value)
				}
			}
		} else {
			// Wildcard is in the middle, need to apply remaining path to all matching keys
			remainingTokens := tokens[wildcardIndex+1:]
			for key, val := range m.AllFromFront() {
				targetNode, ok := val.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", val)
				}
				
				// Construct path for remaining tokens
				remainingPath := ""
				for _, t := range remainingTokens {
					remainingPath += "/" + strings.ReplaceAll(strings.ReplaceAll(t, "~", "~0"), "/", "~1")
				}
				
				// Recursively apply add to this node with remaining path
				err := addValue(targetNode, remainingPath, value)
				if err != nil {
					return fmt.Errorf("failed to add to key %s: %v", key, err)
				}
			}
		}
	} else if current.Kind == core.KindArray {
		arr, ok := current.Value.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", current.Value)
		}
		
		// If this is the last token, add/merge value to all array elements
		if wildcardIndex == len(tokens)-1 {
			// Iterate over all array elements and add/merge the value
			for i, elem := range arr {
				targetNode, ok := elem.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", elem)
				}
				
				// If value is a map, merge it with the existing value
				if value.Kind == core.KindMap && targetNode.Kind == core.KindMap {
					targetMap, ok := targetNode.Value.(*core.YispMap)
					if !ok {
						return fmt.Errorf("expected map, got %T", targetNode.Value)
					}
					valueMap, ok := value.Value.(*core.YispMap)
					if !ok {
						return fmt.Errorf("expected map, got %T", value.Value)
					}
					// Merge value map into target map
					for vKey, vVal := range valueMap.AllFromFront() {
						targetMap.Set(vKey, vVal)
					}
				} else {
					// Replace the entire value
					arr[i] = value
				}
			}
		} else {
			// Wildcard is in the middle, need to apply remaining path to all array elements
			remainingTokens := tokens[wildcardIndex+1:]
			for i, elem := range arr {
				targetNode, ok := elem.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", elem)
				}
				
				// Construct path for remaining tokens
				remainingPath := ""
				for _, t := range remainingTokens {
					remainingPath += "/" + strings.ReplaceAll(strings.ReplaceAll(t, "~", "~0"), "/", "~1")
				}
				
				// Recursively apply add to this node with remaining path
				err := addValue(targetNode, remainingPath, value)
				if err != nil {
					return fmt.Errorf("failed to add to array index %d: %v", i, err)
				}
			}
		}
	} else {
		return fmt.Errorf("wildcard can only be used with maps or arrays, got %s", current.Kind)
	}
	
	return nil
}

// replaceValue replaces a value in a YispNode using a JSON Pointer path (RFC 6902 replace operation)
// Unlike add, this only works if the target path already exists
// Supports wildcard matching with '*' for map keys
func replaceValue(node *core.YispNode, path string, value *core.YispNode) error {
	tokens, err := parsePointer(path)
	if err != nil {
		return err
	}
	
	if len(tokens) == 0 {
		return fmt.Errorf("cannot replace root node")
	}
	
	// Check if path contains wildcard
	wildcardIndex := -1
	for i, token := range tokens {
		if token == "*" {
			wildcardIndex = i
			break
		}
	}
	
	// If wildcard found, handle it specially
	if wildcardIndex >= 0 {
		return replaceValueWithWildcard(node, tokens, wildcardIndex, value)
	}
	
	current := node
	for i := 0; i < len(tokens)-1; i++ {
		token := tokens[i]
		if current.Kind == core.KindMap {
			m, ok := current.Value.(*core.YispMap)
			if !ok {
				return fmt.Errorf("expected map, got %T", current.Value)
			}
			val, ok := m.Get(token)
			if !ok {
				return fmt.Errorf("key not found: %s", token)
			}
			current, ok = val.(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", val)
			}
		} else if current.Kind == core.KindArray {
			arr, ok := current.Value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", current.Value)
			}
			idx, err := strconv.Atoi(token)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", token)
			}
			if idx < 0 || idx >= len(arr) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			current, ok = arr[idx].(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", arr[idx])
			}
		} else {
			return fmt.Errorf("cannot navigate through %s", current.Kind)
		}
	}
	
	lastToken := tokens[len(tokens)-1]
	if current.Kind == core.KindMap {
		m, ok := current.Value.(*core.YispMap)
		if !ok {
			return fmt.Errorf("expected map, got %T", current.Value)
		}
		// Check if key exists
		if _, ok := m.Get(lastToken); !ok {
			return fmt.Errorf("key not found for replace: %s", lastToken)
		}
		m.Set(lastToken, value)
	} else if current.Kind == core.KindArray {
		arr, ok := current.Value.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", current.Value)
		}
		idx, err := strconv.Atoi(lastToken)
		if err != nil {
			return fmt.Errorf("invalid array index: %s", lastToken)
		}
		if idx < 0 || idx >= len(arr) {
			return fmt.Errorf("array index out of bounds: %d", idx)
		}
		arr[idx] = value
	} else {
		return fmt.Errorf("cannot set value in %s", current.Kind)
	}
	
	return nil
}

// replaceValueWithWildcard handles wildcard matching for replace operations
func replaceValueWithWildcard(node *core.YispNode, tokens []string, wildcardIndex int, value *core.YispNode) error {
	// Navigate to the node just before the wildcard
	current := node
	for i := 0; i < wildcardIndex; i++ {
		token := tokens[i]
		if current.Kind == core.KindMap {
			m, ok := current.Value.(*core.YispMap)
			if !ok {
				return fmt.Errorf("expected map, got %T", current.Value)
			}
			val, ok := m.Get(token)
			if !ok {
				return fmt.Errorf("key not found: %s", token)
			}
			current, ok = val.(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", val)
			}
		} else if current.Kind == core.KindArray {
			arr, ok := current.Value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", current.Value)
			}
			idx, err := strconv.Atoi(token)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", token)
			}
			if idx < 0 || idx >= len(arr) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			current, ok = arr[idx].(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", arr[idx])
			}
		} else {
			return fmt.Errorf("cannot navigate through %s", current.Kind)
		}
	}
	
	// At this point, current is the node where wildcard should match
	if current.Kind == core.KindMap {
		m, ok := current.Value.(*core.YispMap)
		if !ok {
			return fmt.Errorf("expected map, got %T", current.Value)
		}
		
		// If this is the last token, replace value for all matching keys
		if wildcardIndex == len(tokens)-1 {
			// Iterate over all keys and replace the value
			for key := range m.AllFromFront() {
				m.Set(key, value)
			}
		} else {
			// Wildcard is in the middle, need to apply remaining path to all matching keys
			remainingTokens := tokens[wildcardIndex+1:]
			for key, val := range m.AllFromFront() {
				targetNode, ok := val.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", val)
				}
				
				// Construct path for remaining tokens
				remainingPath := ""
				for _, t := range remainingTokens {
					remainingPath += "/" + strings.ReplaceAll(strings.ReplaceAll(t, "~", "~0"), "/", "~1")
				}
				
				// Recursively apply replace to this node with remaining path
				err := replaceValue(targetNode, remainingPath, value)
				if err != nil {
					return fmt.Errorf("failed to replace in key %s: %v", key, err)
				}
			}
		}
	} else if current.Kind == core.KindArray {
		arr, ok := current.Value.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", current.Value)
		}
		
		// If this is the last token, replace value for all array elements
		if wildcardIndex == len(tokens)-1 {
			// Iterate over all array elements and replace the value
			for i := range arr {
				arr[i] = value
			}
		} else {
			// Wildcard is in the middle, need to apply remaining path to all array elements
			remainingTokens := tokens[wildcardIndex+1:]
			for i, elem := range arr {
				targetNode, ok := elem.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", elem)
				}
				
				// Construct path for remaining tokens
				remainingPath := ""
				for _, t := range remainingTokens {
					remainingPath += "/" + strings.ReplaceAll(strings.ReplaceAll(t, "~", "~0"), "/", "~1")
				}
				
				// Recursively apply replace to this node with remaining path
				err := replaceValue(targetNode, remainingPath, value)
				if err != nil {
					return fmt.Errorf("failed to replace in array index %d: %v", i, err)
				}
			}
		}
	} else {
		return fmt.Errorf("wildcard can only be used with maps or arrays, got %s", current.Kind)
	}
	
	return nil
}

// deleteValue removes a value from a YispNode using a JSON Pointer path
// Supports wildcard matching with '*' for map keys
func deleteValue(node *core.YispNode, path string) error {
	tokens, err := parsePointer(path)
	if err != nil {
		return err
	}
	
	if len(tokens) == 0 {
		return fmt.Errorf("cannot delete root node")
	}
	
	// Check if path contains wildcard
	wildcardIndex := -1
	for i, token := range tokens {
		if token == "*" {
			wildcardIndex = i
			break
		}
	}
	
	// If wildcard found, handle it specially
	if wildcardIndex >= 0 {
		return deleteValueWithWildcard(node, tokens, wildcardIndex)
	}
	
	current := node
	for i := 0; i < len(tokens)-1; i++ {
		token := tokens[i]
		if current.Kind == core.KindMap {
			m, ok := current.Value.(*core.YispMap)
			if !ok {
				return fmt.Errorf("expected map, got %T", current.Value)
			}
			val, ok := m.Get(token)
			if !ok {
				return fmt.Errorf("key not found: %s", token)
			}
			current, ok = val.(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", val)
			}
		} else if current.Kind == core.KindArray {
			arr, ok := current.Value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", current.Value)
			}
			idx, err := strconv.Atoi(token)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", token)
			}
			if idx < 0 || idx >= len(arr) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			current, ok = arr[idx].(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", arr[idx])
			}
		} else {
			return fmt.Errorf("cannot navigate through %s", current.Kind)
		}
	}
	
	lastToken := tokens[len(tokens)-1]
	if current.Kind == core.KindMap {
		m, ok := current.Value.(*core.YispMap)
		if !ok {
			return fmt.Errorf("expected map, got %T", current.Value)
		}
		if _, ok := m.Get(lastToken); !ok {
			return fmt.Errorf("key not found: %s", lastToken)
		}
		m.Delete(lastToken)
	} else if current.Kind == core.KindArray {
		arr, ok := current.Value.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", current.Value)
		}
		idx, err := strconv.Atoi(lastToken)
		if err != nil {
			return fmt.Errorf("invalid array index: %s", lastToken)
		}
		if idx < 0 || idx >= len(arr) {
			return fmt.Errorf("array index out of bounds: %d", idx)
		}
		// Remove the element at index
		newArr := make([]any, 0, len(arr)-1)
		newArr = append(newArr, arr[:idx]...)
		newArr = append(newArr, arr[idx+1:]...)
		current.Value = newArr
	} else {
		return fmt.Errorf("cannot delete value from %s", current.Kind)
	}
	
	return nil
}

// deleteValueWithWildcard handles wildcard matching for delete operations
func deleteValueWithWildcard(node *core.YispNode, tokens []string, wildcardIndex int) error {
	// Navigate to the node just before the wildcard
	current := node
	for i := 0; i < wildcardIndex; i++ {
		token := tokens[i]
		if current.Kind == core.KindMap {
			m, ok := current.Value.(*core.YispMap)
			if !ok {
				return fmt.Errorf("expected map, got %T", current.Value)
			}
			val, ok := m.Get(token)
			if !ok {
				return fmt.Errorf("key not found: %s", token)
			}
			current, ok = val.(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", val)
			}
		} else if current.Kind == core.KindArray {
			arr, ok := current.Value.([]any)
			if !ok {
				return fmt.Errorf("expected array, got %T", current.Value)
			}
			idx, err := strconv.Atoi(token)
			if err != nil {
				return fmt.Errorf("invalid array index: %s", token)
			}
			if idx < 0 || idx >= len(arr) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			current, ok = arr[idx].(*core.YispNode)
			if !ok {
				return fmt.Errorf("expected YispNode, got %T", arr[idx])
			}
		} else {
			return fmt.Errorf("cannot navigate through %s", current.Kind)
		}
	}
	
	// At this point, current is the node where wildcard should match
	if current.Kind == core.KindMap {
		m, ok := current.Value.(*core.YispMap)
		if !ok {
			return fmt.Errorf("expected map, got %T", current.Value)
		}
		
		// If this is the last token, delete all matching keys
		if wildcardIndex == len(tokens)-1 {
			// Collect all keys first (to avoid modification during iteration)
			keys := []string{}
			for key := range m.AllFromFront() {
				keys = append(keys, key)
			}
			// Delete all keys
			for _, key := range keys {
				m.Delete(key)
			}
		} else {
			// Wildcard is in the middle, need to apply remaining path to all matching keys
			remainingTokens := tokens[wildcardIndex+1:]
			for key, val := range m.AllFromFront() {
				targetNode, ok := val.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", val)
				}
				
				// Construct path for remaining tokens
				remainingPath := ""
				for _, t := range remainingTokens {
					remainingPath += "/" + strings.ReplaceAll(strings.ReplaceAll(t, "~", "~0"), "/", "~1")
				}
				
				// Recursively apply delete to this node with remaining path
				err := deleteValue(targetNode, remainingPath)
				if err != nil {
					return fmt.Errorf("failed to delete from key %s: %v", key, err)
				}
			}
		}
	} else if current.Kind == core.KindArray {
		arr, ok := current.Value.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", current.Value)
		}
		
		// If this is the last token, delete all array elements
		if wildcardIndex == len(tokens)-1 {
			// Clear the array by setting it to empty slice
			current.Value = []any{}
		} else {
			// Wildcard is in the middle, need to apply remaining path to all array elements
			remainingTokens := tokens[wildcardIndex+1:]
			for i, elem := range arr {
				targetNode, ok := elem.(*core.YispNode)
				if !ok {
					return fmt.Errorf("expected YispNode, got %T", elem)
				}
				
				// Construct path for remaining tokens
				remainingPath := ""
				for _, t := range remainingTokens {
					remainingPath += "/" + strings.ReplaceAll(strings.ReplaceAll(t, "~", "~0"), "/", "~1")
				}
				
				// Recursively apply delete to this node with remaining path
				err := deleteValue(targetNode, remainingPath)
				if err != nil {
					return fmt.Errorf("failed to delete from array index %d: %v", i, err)
				}
			}
		}
	} else {
		return fmt.Errorf("wildcard can only be used with maps or arrays, got %s", current.Kind)
	}
	
	return nil
}

func opOpPatch(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("patch requires 2 arguments, got %d", len(cdr)))
	}

	target := cdr[0]
	patchesNode := cdr[1]
	
	// Patches should be an array of patch operations
	if patchesNode.Kind != core.KindArray {
		return nil, core.NewEvaluationError(patchesNode, "patches must be an array")
	}
	
	patchesArray, ok := patchesNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(patchesNode, fmt.Sprintf("expected array, got %T", patchesNode.Value))
	}

	// Apply each patch operation
	for _, patchAny := range patchesArray {
		patchNode, ok := patchAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(patchesNode, fmt.Sprintf("expected YispNode for patch, got %T", patchAny))
		}
		
		if patchNode.Kind != core.KindMap {
			return nil, core.NewEvaluationError(patchNode, "each patch must be a map")
		}
		
		patchMap, ok := patchNode.Value.(*core.YispMap)
		if !ok {
			return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispMap, got %T", patchNode.Value))
		}
		
		// Extract operation type
		opAny, ok := patchMap.Get("op")
		if !ok {
			return nil, core.NewEvaluationError(patchNode, "patch must have 'op' field")
		}
		opNode, ok := opAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispNode for op, got %T", opAny))
		}
		op, ok := opNode.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(opNode, fmt.Sprintf("expected string for op, got %T", opNode.Value))
		}
		
		// Extract path
		pathAny, ok := patchMap.Get("path")
		if !ok {
			return nil, core.NewEvaluationError(patchNode, "patch must have 'path' field")
		}
		pathNode, ok := pathAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispNode for path, got %T", pathAny))
		}
		path, ok := pathNode.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(pathNode, fmt.Sprintf("expected string for path, got %T", pathNode.Value))
		}
		
		switch op {
		case "add":
			// Extract value
			valueAny, ok := patchMap.Get("value")
			if !ok {
				return nil, core.NewEvaluationError(patchNode, "add operation must have 'value' field")
			}
			valueNode, ok := valueAny.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispNode for value, got %T", valueAny))
			}
			
			err := addValue(target, path, valueNode)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("add operation failed: %v", err))
			}
			
		case "remove":
			err := deleteValue(target, path)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("remove operation failed: %v", err))
			}
			
		case "replace":
			// Extract value
			valueAny, ok := patchMap.Get("value")
			if !ok {
				return nil, core.NewEvaluationError(patchNode, "replace operation must have 'value' field")
			}
			valueNode, ok := valueAny.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispNode for value, got %T", valueAny))
			}
			
			err := replaceValue(target, path, valueNode)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("replace operation failed: %v", err))
			}
			
		case "move":
			// Extract from
			fromAny, ok := patchMap.Get("from")
			if !ok {
				return nil, core.NewEvaluationError(patchNode, "move operation must have 'from' field")
			}
			fromNode, ok := fromAny.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispNode for from, got %T", fromAny))
			}
			from, ok := fromNode.Value.(string)
			if !ok {
				return nil, core.NewEvaluationError(fromNode, fmt.Sprintf("expected string for from, got %T", fromNode.Value))
			}
			
			// Get value from 'from' path
			value, err := getValue(target, from)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("move operation (get) failed: %v", err))
			}
			// Remove from 'from' path
			err = deleteValue(target, from)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("move operation (remove) failed: %v", err))
			}
			// Add to 'path'
			err = addValue(target, path, value)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("move operation (add) failed: %v", err))
			}
			
		case "copy":
			// Extract from
			fromAny, ok := patchMap.Get("from")
			if !ok {
				return nil, core.NewEvaluationError(patchNode, "copy operation must have 'from' field")
			}
			fromNode, ok := fromAny.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispNode for from, got %T", fromAny))
			}
			from, ok := fromNode.Value.(string)
			if !ok {
				return nil, core.NewEvaluationError(fromNode, fmt.Sprintf("expected string for from, got %T", fromNode.Value))
			}
			
			// Get value from 'from' path
			value, err := getValue(target, from)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("copy operation (get) failed: %v", err))
			}
			// Add to 'path'
			err = addValue(target, path, value)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("copy operation (add) failed: %v", err))
			}
			
		case "test":
			// Extract value
			valueAny, ok := patchMap.Get("value")
			if !ok {
				return nil, core.NewEvaluationError(patchNode, "test operation must have 'value' field")
			}
			expectedNode, ok := valueAny.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("expected YispNode for value, got %T", valueAny))
			}
			
			// Test that value at path equals specified value
			value, err := getValue(target, path)
			if err != nil {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("test operation failed: %v", err))
			}
			// Compare values (simple comparison for now)
			if value.Kind != expectedNode.Kind {
				return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("test operation failed: kind mismatch"))
			}
			// TODO: Deep comparison if needed
			
		default:
			return nil, core.NewEvaluationError(patchNode, fmt.Sprintf("unknown operation: %s", op))
		}
	}

	return target, nil
}
