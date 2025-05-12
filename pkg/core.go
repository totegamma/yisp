package yisp

import (
	"errors"
	"github.com/totegamma/yisp/yaml"
	"io"
	"os"
)

var (
	allowCmd = false
)

func SetAllowCmd(allow bool) {
	allowCmd = allow
}

func EvaluateYisp(path string) (string, error) {
	env := NewEnv()
	evaluated, err := evaluateYispFile(path, env)
	if err != nil {
		return "", err
	}

	result, err := Render(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func evaluateYispFile(path string, env *Env) (*YispNode, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return evaluateYisp(reader, env, path)
}

func evaluateYisp(document io.Reader, env *Env, location string) (*YispNode, error) {

	decoder := yaml.NewDecoder(document)
	if decoder == nil {
		return nil, errors.New("failed to create decoder")
	}

	documents := make([]any, 0)
	for {
		var root yaml.Node
		err := decoder.Decode(&root)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		parsed, err := Parse(location, &root)
		if err != nil {
			return nil, err
		}

		evaluated, err := Eval(parsed, env, EvalModeQuote)
		if err != nil {
			return nil, err
		}

		if evaluated == nil {
			continue
		}

		if evaluated.Kind == KindNull {
			continue
		}

		documents = append(documents, evaluated)
	}

	return &YispNode{
		Kind:  KindArray,
		Value: documents,
		Tag:   "!expand",
	}, nil
}
