package lib

import (
	"crypto/sha256"
	"fmt"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("crypto", "sha256", opSha256)
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
