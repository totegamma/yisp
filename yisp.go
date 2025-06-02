package yisp

import (
	pkg "github.com/totegamma/yisp/pkg"
)

func SetAllowCmd(allow bool) {
	pkg.SetAllowCmd(allow)
}

func SetShowTrace(show bool) {
	pkg.SetShowTrace(show)
}

func SetAllowedPkgs(pkgs []string) {
	pkg.SetAllowedPkgs(pkgs)
}

func EvaluateYisp(path string) (any, error) {
	env := pkg.NewEnv()
	evaluated, err := pkg.EvaluateYispFile(path, "", env)
	if err != nil {
		return "", err
	}

	result, err := pkg.ToNative(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}

func EvaluateYispToYaml(path string) (string, error) {
	env := pkg.NewEnv()
	evaluated, err := pkg.EvaluateYispFile(path, "", env)
	if err != nil {
		return "", err
	}

	result, err := pkg.Render(evaluated)
	if err != nil {
		return "", err
	}

	return result, nil
}
