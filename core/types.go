package core

import (
	"fmt"
	"github.com/elliotchance/orderedmap/v3"
	"github.com/totegamma/yisp/internal/yaml"
	"os"
	"path/filepath"
)

type YispOperator func([]*YispNode, *Env, EvalMode, Engine) (*YispNode, error)

type YispMap = orderedmap.OrderedMap[string, any]

func NewYispMap() *YispMap {
	return orderedmap.NewOrderedMap[string, any]()
}

type EvalMode int

const (
	EvalModeQuote EvalMode = iota
	EvalModeEval
)

type YamlDocument []any

const (
	YISP_SPECIAL_MERGE_KEY = "__YISP_MERGE_KEY__"
)

// Kind represents the type of a YispNode
type Kind int32

func (k Kind) String() string {
	switch k {
	case KindSymbol:
		return "symbol"
	case KindParameter:
		return "parameter"
	case KindNull:
		return "null"
	case KindBool:
		return "bool"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindString:
		return "string"
	case KindArray:
		return "array"
	case KindMap:
		return "map"
	case KindLambda:
		return "lambda"
	case KindType:
		return "type"
	default:
		return "unknown"
	}
}

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
	KindType
)

type FilePos struct {
	File   string
	Line   int
	Column int
}

type Attribute struct {
	Sources []FilePos

	KeyHeadComment string
	KeyLineComment string
	KeyFootComment string

	HeadComment string
	LineComment string
	FootComment string

	KeyStyle yaml.Style
	Style    yaml.Style
}

func (a Attribute) Merge(other Attribute) Attribute {
	merged := a

	merged.Sources = append(merged.Sources, other.Sources...)

	merged.HeadComment += other.HeadComment
	merged.LineComment += other.LineComment
	merged.FootComment += other.FootComment
	if other.Style != 0 {
		merged.Style = other.Style
	}

	return merged
}

func (a *Attribute) File() string {
	if len(a.Sources) == 0 {
		return ""
	}
	return a.Sources[0].File
}

func (a *Attribute) Line() int {
	if len(a.Sources) == 0 {
		return 0
	}
	return a.Sources[0].Line
}

func (a *Attribute) Column() int {
	if len(a.Sources) == 0 {
		return 0
	}
	return a.Sources[0].Column
}

// YispNode represents a node in the Yisp language
type YispNode struct {
	Kind           Kind
	Tag            string
	Value          any
	Anchor         string
	Attr           Attribute
	IsDocumentRoot bool
	Type           *Schema
}

func (n *YispNode) String() string {
	if len(n.Attr.Sources) == 0 {
		return fmt.Sprintf("%s | <unknown>", n.Kind)
	} else {
		pos := n.Attr.Sources[0]
		return fmt.Sprintf("%s | %s:%d:%d", n.Kind, pos.File, pos.Line, pos.Column)
	}
}

func (n *YispNode) Sourcemap() string {

	if len(n.Attr.Sources) == 0 {
		return ""
	}

	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	output := ""

	for i, pos := range n.Attr.Sources {
		localPath, err := filepath.Rel(currentDir, pos.File)
		if err != nil {
			localPath = pos.File
		}
		if i > 0 {
			output += ", "
		}
		output += fmt.Sprintf("%s:%d:%d", localPath, pos.Line, pos.Column)
	}

	return output
}

type TypedSymbol struct {
	Name   string
	Schema *Schema
}

type Lambda struct {
	Arguments []TypedSymbol
	Returns   *Schema
	Body      *YispNode
	Clojure   *Env
}
