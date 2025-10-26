package lib

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/totegamma/yisp/core"
)

func init() {
	register("exec", "cmd", opCmd)
	register("exec", "go", opGoRun)
}

func opCmd(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("cmdline requires 1 argument, got %d", len(cdr)))
	}

	props := cdr[0]
	if props.Kind != core.KindMap {
		return nil, core.NewEvaluationError(props, fmt.Sprintf("cmdline requires a map argument, got %v", props.Kind))
	}

	propsMap, ok := props.Value.(*core.YispMap)
	if !ok {
		return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid map type: %T", props.Value))
	}

	cmdAny, ok := propsMap.Get("cmd")
	if !ok {
		return nil, core.NewEvaluationError(props, "cmdline requires a 'cmd' key in the map")
	}

	cmdNode, ok := cmdAny.(*core.YispNode)
	if !ok {
		return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid cmd type: %T", cmdAny))
	}
	cmdStr, ok := cmdNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(cmdNode, fmt.Sprintf("invalid cmd value: %T", cmdNode.Value))
	}
	cmd := exec.Command(cmdStr)

	argsAny, ok := propsMap.Get("args")
	if ok {
		argsNode, ok := argsAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid args type: %T", argsAny))
		}

		if argsNode.Kind != core.KindArray {
			return nil, core.NewEvaluationError(argsNode, fmt.Sprintf("args must be an array, got %v", argsNode.Kind))
		}

		arr, ok := argsNode.Value.([]any)
		if !ok {
			return nil, core.NewEvaluationError(argsNode, fmt.Sprintf("invalid array value: %T", argsNode.Value))
		}

		for _, item := range arr {
			node, ok := item.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(argsNode, fmt.Sprintf("invalid item type: %T", item))
			}

			arg, ok := node.Value.(string)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid arg type: %T", node.Value))
			}
			cmd.Args = append(cmd.Args, arg)
		}
	}

	stdinAny, ok := propsMap.Get("stdin")
	if ok {
		stdinNode, ok := stdinAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid stdin type: %T", stdinAny))
		}

		if stdinNode.Kind != core.KindString {
			return nil, core.NewEvaluationError(stdinNode, fmt.Sprintf("stdin must be a string, got %v", stdinNode.Kind))
		}

		str, ok := stdinNode.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(stdinNode, fmt.Sprintf("invalid string value: %T", stdinNode.Value))
		}
		stdin := bytes.NewBufferString(str)
		cmd.Stdin = stdin
	}

	asString := false
	asStringAny, ok := propsMap.Get("asString")
	if ok {
		asStringNode, ok := asStringAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid asString type: %T", asStringAny))
		}
		asString, ok = asStringNode.Value.(bool)
		if !ok {
			return nil, core.NewEvaluationError(asStringNode, fmt.Sprintf("invalid asString value: %T", asStringNode.Value))
		}
	}

	allowCmd := false
	allowCmdAny, ok := e.GetOption("net.gammalab.yisp.exec.allow_cmd")
	if ok {
		allowCmd, ok = allowCmdAny.(bool)
		if !ok {
			return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("invalid net.gammalab.yisp.exec.allowCmd option type: %T", allowCmdAny))
		}
	}

	if !allowCmd {
		fmt.Fprintf(os.Stderr, "Going to run command: %v\n", cmd.Args)
		fmt.Fprintf(os.Stderr, "Press Enter to continue or Ctrl+C to cancel...\n")
		_, err := os.Stdin.Read(make([]byte, 1))
		if err != nil {
			return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("failed to read input: %v", err))
		}
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	errorOutput := stderr.String()
	if err != nil {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("command execution error: %s", errorOutput))
	}

	if asString {
		return &core.YispNode{
			Kind:  core.KindString,
			Value: stdout.String(),
		}, nil
	} else {

		result, err := e.Run(stdout, env, cdr[0].Attr.File())
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(cdr[0], "failed to evaluate command output", err)
		}

		return result, nil
	}
}

func opGoRun(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("gorun requires 1 argument, got %d", len(cdr)))
	}
	props := cdr[0]
	if props.Kind != core.KindMap {
		return nil, core.NewEvaluationError(props, fmt.Sprintf("gorun requires a map argument, got %v", props.Kind))
	}
	propsMap, ok := props.Value.(*core.YispMap)
	if !ok {
		return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid map type: %T", props.Value))
	}

	pkgAny, ok := propsMap.Get("pkg")
	if !ok {
		return nil, core.NewEvaluationError(props, "gorun requires a 'pkg' key in the map")
	}

	pkgNode, ok := pkgAny.(*core.YispNode)
	if !ok {
		return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid pkg type: %T", pkgAny))
	}
	pkgStr, ok := pkgNode.Value.(string)
	if !ok {
		return nil, core.NewEvaluationError(pkgNode, fmt.Sprintf("invalid pkg value: %T", pkgNode.Value))
	}

	allowedGoPkgs := []string{}
	allowedGoPkgsAny, ok := e.GetOption("net.gammalab.yisp.exec.allowed_go_pkgs")
	if ok {
		allowedGoPkgsSlice, ok := allowedGoPkgsAny.([]string)
		if !ok {
			return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("invalid net.gammalab.yisp.exec.allowedGoPkgs option type: %T", allowedGoPkgsAny))
		}
		for _, item := range allowedGoPkgsSlice {
			allowedGoPkgs = append(allowedGoPkgs, item)
		}
	}

	allowed := false
	for _, stmt := range allowedGoPkgs {
		regex := "^" + strings.ReplaceAll(regexp.QuoteMeta(stmt), "\\*", ".*") + "$"
		matched, err := regexp.MatchString(regex, pkgStr)
		if err != nil {
			return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("failed to match package %s with regex %s: %v", pkgStr, regex, err))
		}
		if matched {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("package %s is not allowed. Run command below to allow it:\n\nyisp allow %s", pkgStr, pkgStr))
	}

	cmd := exec.Command("go", "run", pkgStr)

	argsAny, ok := propsMap.Get("args")
	if ok {
		argsNode, ok := argsAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid args type: %T", argsAny))
		}

		if argsNode.Kind != core.KindArray {
			return nil, core.NewEvaluationError(argsNode, fmt.Sprintf("args must be an array, got %v", argsNode.Kind))
		}

		arr, ok := argsNode.Value.([]any)
		if !ok {
			return nil, core.NewEvaluationError(argsNode, fmt.Sprintf("invalid array value: %T", argsNode.Value))
		}

		for _, item := range arr {
			node, ok := item.(*core.YispNode)
			if !ok {
				return nil, core.NewEvaluationError(argsNode, fmt.Sprintf("invalid item type: %T", item))
			}
			arg, ok := node.Value.(string)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid arg type: %T", node.Value))
			}
			cmd.Args = append(cmd.Args, arg)
		}
	}

	stdinAny, ok := propsMap.Get("stdin")
	if ok {
		stdinNode, ok := stdinAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid stdin type: %T", stdinAny))
		}
		str, ok := stdinNode.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(stdinNode, fmt.Sprintf("invalid string value: %T", stdinNode.Value))
		}
		stdin := bytes.NewBufferString(str)
		cmd.Stdin = stdin
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	errorOutput := stderr.String()
	if err != nil {
		return nil, core.NewEvaluationError(cdr[0], fmt.Sprintf("command execution error: %s", errorOutput))
	}

	asString := false
	asStringAny, ok := propsMap.Get("asString")
	if ok {
		asStringNode, ok := asStringAny.(*core.YispNode)
		if !ok {
			return nil, core.NewEvaluationError(props, fmt.Sprintf("invalid asString type: %T", asStringAny))
		}
		asString, ok = asStringNode.Value.(bool)
		if !ok {
			return nil, core.NewEvaluationError(asStringNode, fmt.Sprintf("invalid asString value: %T", asStringNode.Value))
		}
	}

	if asString {
		return &core.YispNode{
			Kind:  core.KindString,
			Value: stdout.String(),
		}, nil
	} else {

		result, err := e.Run(stdout, env, cdr[0].Attr.File())
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(cdr[0], "failed to evaluate command output", err)
		}

		return result, nil
	}
}
