package yisp

import (
	"maps"
	"strings"
)

// Env represents the execution environment with variable bindings
type Env struct {
	Parent *Env
	Vars   map[string]*YispNode
}

// NewEnv creates a new environment with an empty variable map
func NewEnv() *Env {
	return &Env{
		Vars: make(map[string]*YispNode),
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
		Parent: e.Parent,
		Vars:   make(map[string]*YispNode),
	}
	maps.Copy(clone.Vars, e.Vars)
	return clone
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

	split := strings.Split(key, ".")
	value, ok := e.Vars[split[0]]
	if !ok {
		if e.Parent != nil {
			return e.Parent.Get(key)
		}
		return nil, false
	}

	for _, key := range split[1:] {

		maps, ok := value.Value.(map[string]*YispNode)
		if !ok {
			anyMaps, ok := value.Value.(map[string]any)
			if !ok {
				return nil, false
			}
			maps = make(map[string]*YispNode)
			for key, item := range anyMaps {
				node, ok := item.(*YispNode)
				if !ok {
					continue
				}
				maps[key] = node
			}
		}

		value, ok = maps[key]
		if !ok {
			return nil, false
		}
	}

	return value, ok
}
