package yisp

import (
	"fmt"
	"maps"
	"strings"
)

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

func (e *Env) Depth() int {
	depth := 0
	for env := e; env != nil; env = env.Parent {
		depth++
	}
	return depth
}

func (e *Env) Clone() *Env {
	clone := &Env{
		Parent:  e.Parent,
		Vars:    make(map[string]*YispNode),
		Modules: make(map[string]*Env),
	}
	maps.Copy(clone.Vars, e.Vars)
	maps.Copy(clone.Modules, e.Modules)
	return clone
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
