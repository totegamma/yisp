package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/internal/yaml"
)

type engine struct {
	execOptions          map[string]any
	showTrace            bool
	renderSources        bool
	renderSpecialObjects bool
	allowUntypedManifest bool
}

type Options struct {
	ShowTrace            bool
	RenderSources        bool
	RenderSpecialObjects bool
	AllowUntypedManifest bool
}

func NewEngine(opts Options) *engine {
	return &engine{
		execOptions:          make(map[string]any),
		showTrace:            opts.ShowTrace,
		renderSpecialObjects: opts.RenderSpecialObjects,
		renderSources:        opts.RenderSources,
		allowUntypedManifest: opts.AllowUntypedManifest,
	}
}

func (e *engine) SetOption(key string, value any) {
	e.execOptions[key] = value
}

func (e *engine) GetOption(key string) (any, bool) {
	if value, ok := e.execOptions[key]; ok {
		return value, true
	}
	return nil, false
}

func (e *engine) EvaluateFileToYamlWithEnv(path string, env *core.Env) (string, error) {
	evaluated, err := core.CallEngineByPath(path, "", env, e)
	if err != nil {
		return "", err
	}

	result, err := e.Render(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (e *engine) EvaluateFileToYaml(path string) (string, error) {
	env := core.NewEnv()
	return e.EvaluateFileToYamlWithEnv(path, env)
}

func (e *engine) EvaluateReaderToYamlWithEnv(reader io.Reader, env *core.Env, location string) (string, error) {
	evaluated, err := e.Run(reader, env, location)
	if err != nil {
		return "", err
	}

	result, err := e.Render(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (e *engine) EvaluateReaderToYaml(reader io.Reader, location string) (string, error) {
	env := core.NewEnv()
	return e.EvaluateReaderToYamlWithEnv(reader, env, location)
}

func (e *engine) EvaluateFileToAny(path string) (any, error) {
	env := core.NewEnv()
	evaluated, err := core.CallEngineByPath(path, "", env, e)
	if err != nil {
		return "", err
	}

	result, err := ToNative(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (e *engine) EvaluateBytesToYaml(data []byte, global map[string]any) (string, error) {
	env := core.NewEnv()

	for key, value := range global {
		node, err := core.ParseAny("", value)
		if err != nil {
			return "", fmt.Errorf("failed to parse global variable %s: %v", key, err)
		}
		env.Set(key, node)
	}

	reader := io.NopCloser(bytes.NewReader(data))
	evaluated, err := e.Run(reader, env, "inline")
	if err != nil {
		return "", err
	}

	result, err := e.Render(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (e *engine) Run(document io.Reader, env *core.Env, location string) (*core.YispNode, error) {

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

		evaluated, err := e.Eval(parsed, env, core.EvalModeQuote)
		if err != nil {
			return nil, err
		}

		if evaluated == nil {
			continue
		}

		if evaluated.Kind == core.KindNull {
			continue
		}

		if evaluated.Kind == core.KindArray && evaluated.IsDocumentRoot {
			arr, ok := evaluated.Value.([]any)
			if !ok {
				return nil, fmt.Errorf("invalid array value")
			}
			documents = append(documents, arr...)
			continue
		} else {
			documents = append(documents, evaluated)
		}
	}

	return &core.YispNode{
		Kind:           core.KindArray,
		Value:          documents,
		IsDocumentRoot: true,
	}, nil

}
