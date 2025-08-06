package core

import (
	"io"
)

type Engine interface {
	Run(document io.Reader, env *Env, location string) (*YispNode, error)
	Apply(car *YispNode, cdr []*YispNode, env *Env, mode EvalMode) (*YispNode, error)
	Eval(node *YispNode, env *Env, mode EvalMode) (*YispNode, error)
	Render(node *YispNode) (string, error)
	GetOption(key string) (any, bool)
}
