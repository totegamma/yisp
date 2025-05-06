package yisp

import (
	"maps"
)

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

// Environment represents the execution environment with variable bindings
type Environment struct {
	Vars map[string]*YispNode
}

// NewEnvironment creates a new environment with an empty variable map
func NewEnvironment() *Environment {
	return &Environment{
		Vars: make(map[string]*YispNode),
	}
}

// Clone creates a copy of the environment
func (env *Environment) Clone() *Environment {
	newEnv := NewEnvironment()
	maps.Copy(newEnv.Vars, env.Vars)
	return newEnv
}

// Globals is the global environment for storing anchors
var globals = NewEnvironment()

// GetGlobals returns the global environment
func GetGlobals() *Environment {
	return globals
}
