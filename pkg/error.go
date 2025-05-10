package yisp

import (
	"fmt"
)

var EvaluationError = ErrorTypeEvaluation{Type: "YispErrorEvaluation"}

type ErrorTypeEvaluation struct {
	Type    string
	Node    *YispNode
	Message string
}

func NewEvaluationError(node *YispNode, message string) *ErrorTypeEvaluation {
	return &ErrorTypeEvaluation{
		Type:    "YispErrorEvaluation",
		Node:    node,
		Message: message,
	}
}

func (e *ErrorTypeEvaluation) Error() string {
	file := ""
	line := 0
	column := 0

	if e.Node != nil {
		file = e.Node.File
		line = e.Node.Line
		column = e.Node.Column
	}

	return fmt.Sprintf("\nEvaluation error at %s:%d:%d %s", file, line, column, e.Message)
}
