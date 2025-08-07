package lib

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("strings", "concat", opConcat)
	register("strings", "contains", opContains)
	register("strings", "escape", opEscape)
	register("strings", "format", opFormat)
	register("strings", "join", opJoin)
	register("strings", "replace", opReplace)
	register("strings", "sha256", opSha256)
	register("strings", "split", opSplit)
	register("strings", "toUpper", opToUpper)
	register("strings", "toLower", opToLower)
	register("strings", "trim", opTrim)
}

// opConcat concatenates strings
func opConcat(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	var result string
	for _, node := range cdr {
		str, ok := node.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for concat: %T", node))
		}
		result += str
	}

	return &core.YispNode{
		Kind:  core.KindString,
		Value: result,
	}, nil
}

// contains
func opContains(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("contains requires 2 arguments, got %d", len(cdr)))
	}

	strNode := cdr[0]
	substrNode := cdr[1]

	str, ok := strNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(strNode, fmt.Sprintf("contains requires a string to search in, got %v", strNode.Kind))
	}

	substr, ok := substrNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(substrNode, fmt.Sprintf("contains requires a string to search for, got %v", substrNode.Kind))
	}

	return &core.YispNode{
		Kind:  core.KindBool,
		Value: strings.Contains(str, substr),
	}, nil
}

func opEscape(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("escape requires 1 argument, got %d", len(cdr)))
	}

	value := ""
	switch cdr[0].Kind {
	case core.KindString:
		value, _ = cdr[0].Value.(string)
	case core.KindInt:
		value = fmt.Sprintf("%d", cdr[0].Value)
	case core.KindFloat:
		value = fmt.Sprintf("%f", cdr[0].Value)
	case core.KindBool:
		if cdr[0].Value.(bool) {
			value = "true"
		} else {
			value = "false"
		}
	default:
		value = fmt.Sprintf("%v", cdr[0].Value)
	}

	node := cdr[0]
	str := fmt.Sprintf("%q", value)
	str = strings.Trim(str, "\"")
	str = strings.Trim(str, "'")

	return &core.YispNode{
		Kind:  core.KindString,
		Value: str,
		Attr:  node.Attr,
	}, nil
}

func opFormat(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	formatNode := cdr[0]
	argsNode := cdr[1:]

	formatStr, ok := formatNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(formatNode, fmt.Sprintf("format requires a string argument, got %v", formatNode.Kind))
	}

	args := make([]any, len(argsNode))
	for i, arg := range argsNode {
		args[i] = arg.Value
	}

	return &core.YispNode{
		Kind:  core.KindString,
		Value: fmt.Sprintf(formatStr, args...),
		Attr:  formatNode.Attr,
	}, nil

}

func opJoin(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("join requires 2 arguments, got %d", len(cdr)))
	}

	strsNode := cdr[0]
	sepNode := cdr[1]

	sep, ok := sepNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(sepNode, fmt.Sprintf("join requires a string separator, got %v", sepNode.Kind))
	}

	strsAny, ok := strsNode.Value.([]any)
	if !ok {
		return nil, core.NewEvaluationError(strsNode, fmt.Sprintf("join requires a list of strings, got %v", strsNode.Kind))
	}

	strs := make([]string, len(strsAny))
	for i, item := range strsAny {
		strNode, ok := item.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(strsNode, fmt.Sprintf("join requires a list of strings, got %T", item))
		}
		if strNode.Kind != core.KindString {
			return nil, core.NewEvaluationError(strNode, fmt.Sprintf("join requires a string, got %v", strNode.Kind))
		}
		str, ok := strNode.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(strNode, fmt.Sprintf("join requires a string, got %T", strNode.Value))
		}
		strs[i] = str
	}

	result := strings.Join(strs, sep)

	return &core.YispNode{
		Kind:  core.KindString,
		Value: result,
	}, nil
}

func opReplace(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) < 3 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("replace requires 3-4 arguments, got %d", len(cdr)))
	}

	n := -1

	strNode := cdr[0]
	oldNode := cdr[1]
	newNode := cdr[2]

	if len(cdr) == 4 {
		nNode := cdr[3]
		if nNode.Kind != core.KindInt {
			return nil, core.NewEvaluationError(nNode, fmt.Sprintf("replace requires an integer for the count, got %v", nNode.Kind))
		}
		var ok bool
		n, ok = nNode.Value.(int)
		if !ok {
			return nil, core.NewEvaluationError(nNode, fmt.Sprintf("replace requires an integer for the count, got %T", nNode.Value))
		}
	}

	str, ok := strNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(strNode, fmt.Sprintf("replace requires a string to operate on, got %v", strNode.Kind))
	}

	oldStr, ok := oldNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(oldNode, fmt.Sprintf("replace requires a string to replace, got %v", oldNode.Kind))
	}

	newStr, ok := newNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(newNode, fmt.Sprintf("replace requires a string to replace with, got %v", newNode.Kind))
	}

	result := strings.Replace(str, oldStr, newStr, n)

	return &core.YispNode{
		Kind:  core.KindString,
		Value: result,
	}, nil
}

func opSha256(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("sha256 requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	if node.Kind != core.KindString {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("sha256 requires a string argument, got %v", node.Kind))
	}

	str, ok := node.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for sha256: %T", node.Value))
	}

	hash := sha256.Sum256([]byte(str))

	return &core.YispNode{
		Kind:  core.KindString,
		Value: fmt.Sprintf("%x", hash),
		Attr:  node.Attr,
	}, nil
}

func opSplit(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 2 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("split requires 2 arguments, got %d", len(cdr)))
	}

	strNode := cdr[0]
	sepNode := cdr[1]

	str, ok := strNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(strNode, fmt.Sprintf("split requires a string to split, got %v", strNode.Kind))
	}

	sep, ok := sepNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(sepNode, fmt.Sprintf("split requires a string separator, got %v", sepNode.Kind))
	}

	result := strings.Split(str, sep)

	// Convert the result to a slice of YispNode
	var nodes []any
	for _, item := range result {
		nodes = append(nodes, &core.YispNode{
			Kind:  core.KindString,
			Value: item,
			Attr:  strNode.Attr, // Preserve attributes from the original string node
		})
	}

	return &core.YispNode{
		Kind:  core.KindArray,
		Value: nodes,
	}, nil
}

func opToUpper(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("toUpper requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]
	if node.Kind != core.KindString {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("toUpper requires a string argument, got %v", node.Kind))
	}

	str, ok := node.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for toUpper: %T", node.Value))
	}

	return &core.YispNode{
		Kind:  core.KindString,
		Value: strings.ToUpper(str),
		Attr:  node.Attr,
	}, nil
}

func opToLower(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("toLower requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]
	if node.Kind != core.KindString {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("toLower requires a string argument, got %v", node.Kind))
	}

	str, ok := node.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for toLower: %T", node.Value))
	}

	return &core.YispNode{
		Kind:  core.KindString,
		Value: strings.ToLower(str),
		Attr:  node.Attr,
	}, nil
}

func opTrim(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("trim requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]
	if node.Kind != core.KindString {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("trim requires a string argument, got %v", node.Kind))
	}

	str, ok := node.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for trim: %T", node.Value))
	}

	return &core.YispNode{
		Kind:  core.KindString,
		Value: strings.TrimSpace(str),
		Attr:  node.Attr,
	}, nil
}
