package lib

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("files", "glob", opReadFiles)
}

func opReadFiles(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	result := make([]any, 0)

	for _, node := range cdr {
		str, ok := node.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for open: %T", node))
		}

		path := str
		if node.Attr.File != "" {
			path = filepath.Clean(filepath.Join(filepath.Dir(node.Attr.File), str))
		}

		files, err := filepath.Glob(path)
		if err != nil {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("failed to glob path: %s", str))
		}

		for _, file := range files {

			filename := filepath.Base(file)
			body, err := os.ReadFile(file)
			if err != nil {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("failed to read file: %s", file))
			}

			value := core.NewYispMap()
			value.Set("path", &core.YispNode{
				Kind:  core.KindString,
				Value: file,
			})
			value.Set("name", &core.YispNode{
				Kind:  core.KindString,
				Value: filename,
			})
			value.Set("body", &core.YispNode{
				Kind:  core.KindString,
				Value: string(body),
			})

			result = append(result, &core.YispNode{
				Kind:  core.KindMap,
				Value: value,
				Attr: core.Attribute{
					File:   file,
					Line:   node.Attr.Line,
					Column: node.Attr.Column,
				},
			})
		}
	}

	return &core.YispNode{
		Kind:  core.KindArray,
		Value: result,
		Attr:  cdr[0].Attr,
	}, nil
}
