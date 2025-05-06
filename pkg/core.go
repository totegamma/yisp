package yisp

import (
	"errors"
	"github.com/totegamma/yisp/yaml"
	"io"
	"os"
)

func EvaluateYisp(path string) (string, error) {
	env := NewEnv()
	evaluated, err := evaluateYisp(path, env)
	if err != nil {
		return "", err
	}

	result, err := Render(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func evaluateYisp(path string, env *Env) (*YispNode, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoder := yaml.NewDecoder(reader)
	if decoder == nil {
		return nil, errors.New("failed to create decoder")
	}

	documents := make([]*YispNode, 0)
	for {
		var root yaml.Node
		err := decoder.Decode(&root)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		parsed, err := Parse(path, &root, env)
		if err != nil {
			return nil, err
		}

		evaluated, err := Eval(parsed, env)
		if err != nil {
			return nil, err
		}

		if evaluated == nil {
			continue
		}

		documents = append(documents, evaluated)
	}

	return &YispNode{
		Kind:  KindArray,
		Value: documents,
	}, nil
}
