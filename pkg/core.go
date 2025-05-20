package yisp

import (
	"errors"
	"fmt"
	"github.com/totegamma/yisp/yaml"
	"io"
	"net/http"
	"net/url"
	"os"
)

var (
	allowCmd      = false
	showTrace     = false
	allowedGoPkgs = []string{}
)

func SetAllowCmd(allow bool) {
	allowCmd = allow
}

func SetShowTrace(show bool) {
	showTrace = show
}

func SetAllowedPkgs(pkgs []string) {
	allowedGoPkgs = pkgs
}

func EvaluateYisp(path string) (string, error) {
	env := NewEnv()
	evaluated, err := evaluateYispFile(path, "", env)
	if err != nil {
		return "", err
	}

	result, err := Render(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func evaluateYispFile(path, base string, env *Env) (*YispNode, error) {

	var reader io.Reader
	var err error

	targetURL, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	if base != "" {
		baseURL, err := url.Parse(base)
		if err != nil {
			return nil, fmt.Errorf("failed to parse base URL: %v", err)
		}
		targetURL = baseURL.ResolveReference(targetURL)
	}

	if targetURL.Scheme == "http" || targetURL.Scheme == "https" {
		reader, err = fetchRemote(targetURL.String())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch remote file: %v", err)
		}
	} else {
		reader, err = os.Open(targetURL.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
	}

	return evaluateYisp(reader, env, targetURL.String())
}

func fetchRemote(rawURL string) (io.ReadCloser, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote file: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to fetch remote file: %s", resp.Status)
	}
	return resp.Body, nil
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
