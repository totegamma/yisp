package yisp

import (
	"fmt"
	"strings"
)

type EvalMode int

const (
	EvalModeQuote EvalMode = iota
	EvalModeEval
)

type YamlDocument []any

// Kind represents the type of a YispNode
type Kind int32

func (k Kind) String() string {
	switch k {
	case KindSymbol:
		return "symbol"
	case KindParameter:
		return "parameter"
	case KindNull:
		return "null"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindString:
		return "string"
	case KindArray:
		return "array"
	case KindMap:
		return "map"
	case KindLambda:
		return "lambda"
	default:
		return "unknown"
	}
}

const (
	KindSymbol Kind = iota
	KindParameter
	KindNull
	KindBool
	KindInt
	KindFloat
	KindString
	KindArray
	KindMap
	KindLambda
)

// YispNode represents a node in the Yisp language
type YispNode struct {
	Kind   Kind
	Tag    string
	Value  any
	File   string
	Line   int
	Column int
	Anchor string
}

// Env represents the execution environment with variable bindings
type Env struct {
	Parent  *Env
	Vars    map[string]*YispNode
	Modules map[string]*Env
}

// NewEnv creates a new environment with an empty variable map
func NewEnv() *Env {
	return &Env{
		Vars:    make(map[string]*YispNode),
		Modules: make(map[string]*Env),
	}
}

func (e *Env) CreateChild() *Env {
	return &Env{
		Parent:  e,
		Vars:    make(map[string]*YispNode),
		Modules: make(map[string]*Env),
	}
}

func (e *Env) Set(key string, value *YispNode) {
	e.Vars[key] = value
}

func (e *Env) Get(key string) (*YispNode, bool) {

	split := strings.Split(key, "__")
	if len(split) > 1 {
		moduleName := split[0]
		resolvedName := strings.Join(split[1:], "__")

		if module, ok := e.Modules[moduleName]; ok {
			return module.Get(resolvedName)
		} else {
			fmt.Println("Module not found:", moduleName)
			return nil, false
		}
	}

	if value, ok := e.Vars[key]; ok {
		return value, true
	}
	if e.Parent != nil {
		return e.Parent.Get(key)
	}
	return nil, false
}

func (e *Env) AddModule(name string, module *Env) {
	e.Modules[name] = module
}
