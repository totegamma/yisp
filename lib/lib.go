package lib

import (
	"github.com/totegamma/yisp/core"
)

type moduleOperator struct {
	Module string
	Name   string
}

var operators map[moduleOperator]core.YispOperator

func register(module, name string, op core.YispOperator) {
	if operators == nil {
		operators = make(map[moduleOperator]core.YispOperator)
	}
	operators[moduleOperator{Module: module, Name: name}] = op
}

func GetOperator(module, name string) (core.YispOperator, bool) {
	op, ok := operators[moduleOperator{Module: module, Name: name}]
	return op, ok
}
