package lib

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("strings", "format", opFormat)
	register("strings", "escape", opEscape)
	register("strings", "concat", opConcat)
	register("strings", "sha256", opSha256)
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
