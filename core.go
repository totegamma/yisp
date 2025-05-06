package yisp

import (
	"errors"
	"github.com/totegamma/yisp/yaml"
	"io"
	"os"
)

func EvaluateYisp(path string, parent *Environment) *YispNode {
	reader, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	decoder := yaml.NewDecoder(reader)
	if decoder == nil {
		panic("failed to create decoder")
	}

	documents := make([]*YispNode, 0)
	for {
		var root yaml.Node
		err := decoder.Decode(&root)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}

		env := parent.Clone()
		parsed, err := Parse(&root, env)
		if err != nil {
			panic(err)
		}

		evaluated, err := Eval(parsed, GetGlobals())
		if err != nil {
			panic(err)
		}

		if evaluated == nil {
			continue
		}

		documents = append(documents, evaluated)
	}

	return &YispNode{
		Kind:  KindArray,
		Value: documents,
	}
}
