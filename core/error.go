package core

import (
	"fmt"
)

var EvaluationError = ErrorTypeEvaluation{Type: "YispErrorEvaluation"}

type ErrorTypeEvaluation struct {
	Type    string
	Node    *YispNode
	Message string
	Parent  *ErrorTypeEvaluation
}

func NewEvaluationError(node *YispNode, message string) *ErrorTypeEvaluation {
	return &ErrorTypeEvaluation{
		Type:    "YispErrorEvaluation",
		Node:    node,
		Message: message,
	}
}

func NewEvaluationErrorWithParent(node *YispNode, message string, parent error) *ErrorTypeEvaluation {

	p, ok := parent.(*ErrorTypeEvaluation)
	if ok {
		return &ErrorTypeEvaluation{
			Type:    "YispErrorEvaluation",
			Node:    node,
			Message: message,
			Parent:  p,
		}
	} else {
		return &ErrorTypeEvaluation{
			Type:    "YispErrorEvaluation",
			Node:    node,
			Message: message + " (" + parent.Error() + ")",
			Parent:  nil,
		}
	}
}

func (e *ErrorTypeEvaluation) String() string {
	file := ""
	line := 0
	column := 0

	if e.Node != nil {
		file = e.Node.Attr.File
		line = e.Node.Attr.Line
		column = e.Node.Attr.Column
	}

	return fmt.Sprintf("%s at %s:%d:%d", e.Message, file, line, column)
}

func (e *ErrorTypeEvaluation) GetRoot() *ErrorTypeEvaluation {
	if e.Parent != nil {
		return e.Parent.GetRoot()
	}
	return e
}

func (e *ErrorTypeEvaluation) Error() string {

	if e.Parent == nil { // root cause
		message := e.Message
		message += "\n"

		if e.Node != nil {
			message += "\n"

			line, err := RenderCode(e.Node.Attr.File, e.Node.Attr.Line, 3, 3, []Comment{
				{
					Line:   e.Node.Attr.Line,
					Column: e.Node.Attr.Column,
					Text:   e.Message,
				},
			})
			if err != nil {
				message += fmt.Sprintf("Error reading file: %s\n", err)
			} else {
				message += fmt.Sprintf("%s\n", line)
			}

		}

		message += "Traceback:\n"
		message += e.String()

		return message

	} else {

		return e.Parent.Error() + "\n" + e.String()
	}
}

type Comment struct {
	Line   int
	Column int
	Text   string
}
