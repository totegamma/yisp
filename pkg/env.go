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
		Parent: nil,
		Vars: map[string]*YispNode{
			"null": {
				Kind: KindType,
				Value: &Schema{
					Type: "null",
				},
			},
			"bool": {
				Kind: KindType,
				Value: &Schema{
					Type: "boolean",
				},
			},
			"int": {
				Kind: KindType,
				Value: &Schema{
					Type: "integer",
				},
			},
			"float": {
				Kind: KindType,
				Value: &Schema{
					Type: "float",
				},
			},
			"string": {
				Kind: KindType,
				Value: &Schema{
					Type: "string",
				},
			},
		},
	}
}

func (e *Env) Root() *Env {
	if e.Parent == nil {
		return e
	}
	return e.Parent.Root()
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

	fst := split[0]
	optional := false
	if fst[len(fst)-1] == '?' {
		fst = fst[:len(fst)-1]
		optional = true
	}

	value, ok := e.Vars[split[0]]
	if !ok {
		if e.Parent != nil {
			return e.Parent.Get(key)
		}
		if optional {
			return &YispNode{
				Kind:  KindNull,
				Value: nil,
			}, true
		}
		return nil, false
	}

	for _, key := range split[1:] {
		if key[len(key)-1] == '?' {
			key = key[:len(key)-1]
			optional = true
		} else {
			optional = false
		}

		maps, ok := value.Value.(map[string]*YispNode)
		if !ok {
			anyMaps, ok := value.Value.(*YispMap)
			if !ok {
				return nil, false
			}
			maps = make(map[string]*YispNode)
			for key, item := range anyMaps.AllFromFront() {
				node, ok := item.(*YispNode)
				if !ok {
					continue
				}
				maps[key] = node
			}
		}

		value, ok = maps[key]
		if !ok {
			if optional {
				return &YispNode{
					Kind:  KindNull,
					Value: nil,
				}, true
			} else {
				return nil, false
			}
		}
	}

	return value, ok
}
