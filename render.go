package yisp

// Render converts a YispNode to a native Go value
func Render(node *YispNode) any {
	switch node.Kind {
	case KindNull, KindBool, KindInt, KindFloat, KindString:
		return node.Value
	case KindArray:
		arr, ok := node.Value.([]any)
		if !ok {
			return nil
		}
		results := make([]any, len(arr))
		for i, item := range arr {
			node, ok := item.(*YispNode)
			if !ok {
				return nil
			}
			results[i] = Render(node)
		}
		return results
	case KindMap:
		m, ok := node.Value.(map[string]any)
		if !ok {
			return nil
		}
		results := make(map[string]any)
		for key, item := range m {
			node, ok := item.(*YispNode)
			if !ok {
				return nil
			}
			results[key] = Render(node)
		}
		return results
	case KindLambda:
		return "(lambda)"
	case KindParameter:
		return "(parameter)"
	case KindSymbol:
		return "(symbol)"
	default:
		return "(unknown)"
	}
}
