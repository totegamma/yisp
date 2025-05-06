package yisp

type YamlDocument []any

// Kind represents the type of a YispNode
type Kind int32

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
	Kind  Kind
	Tag   string
	Value any
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
		Vars: make(map[string]*YispNode),
	}
}

func (e *Env) CreateChild() *Env {
	return &Env{
		Parent: e,
		Vars:   make(map[string]*YispNode),
	}
}

func (e *Env) Set(key string, value *YispNode) {
	e.Vars[key] = value
}

func (e *Env) Get(key string) (*YispNode, bool) {
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
