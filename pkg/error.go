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
	return fmt.Sprintf("\nEvaluation error at %s:%d:%d %s", e.Node.File, e.Node.Line, e.Node.Column, e.Message)
}
