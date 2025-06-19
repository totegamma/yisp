package yisp

import (
	"fmt"
	"github.com/elliotchance/orderedmap/v3"
	"github.com/totegamma/yisp/internal/yaml"
)

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

type Attribute struct {
	File   string
	Line   int
	Column int

	KeyHeadComment string
	KeyLineComment string
	KeyFootComment string

	HeadComment string
	LineComment string
	FootComment string

	KeyStyle yaml.Style
	Style    yaml.Style
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
	return fmt.Sprintf("%s | %s:%d:%d", n.Kind, n.Attr.File, n.Attr.Line, n.Attr.Column)
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
