package core

import "strings"

// LookupYispNodeByPath resolves dot-separated map paths like "metadata.name".
// A path segment ending in "?" returns null when that segment is missing.
func LookupYispNodeByPath(root any, path string) (*YispNode, bool) {
	return lookupYispNodeByPathSegments(root, strings.Split(path, "."))
}

func lookupYispNodeByPathSegments(root any, segments []string) (*YispNode, bool) {
	if len(segments) == 0 {
		node, ok := root.(*YispNode)
		return node, ok
	}

	current := root
	for _, segment := range segments {
		key, optional := parsePathSegment(segment)
		if key == "" {
			return nil, false
		}

		node, ok := lookupYispNodeChild(current, key)
		if !ok {
			if optional {
				return newNullYispNode(), true
			}
			return nil, false
		}
		current = node
	}

	node, ok := current.(*YispNode)
	return node, ok
}

func parsePathSegment(segment string) (string, bool) {
	if strings.HasSuffix(segment, "?") {
		return strings.TrimSuffix(segment, "?"), true
	}
	return segment, false
}

func lookupYispNodeChild(root any, key string) (*YispNode, bool) {
	switch value := root.(type) {
	case map[string]*YispNode:
		node, ok := value[key]
		return node, ok
	case *YispMap:
		item, ok := value.Get(key)
		if !ok {
			return nil, false
		}
		node, ok := item.(*YispNode)
		return node, ok
	case *YispNode:
		return lookupYispNodeChild(value.Value, key)
	default:
		return nil, false
	}
}

func newNullYispNode() *YispNode {
	return &YispNode{
		Kind:  KindNull,
		Value: nil,
	}
}
