package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/totegamma/yisp/core"
	"github.com/totegamma/yisp/lib"
)

// Apply applies a function to arguments
func (e *engine) Apply(car *core.YispNode, cdr []*core.YispNode, env *core.Env, mode core.EvalMode) (*core.YispNode, error) {

	switch car.Kind {
	case core.KindLambda:
		lambda, ok := car.Value.(*core.Lambda)
		if !ok {
			return nil, core.NewEvaluationError(car, fmt.Sprintf("invalid lambda type: %T", car.Value))
		}

		newEnv := lambda.Clojure.CreateChild()
		for i, node := range cdr {
			if lambda.Arguments[i].Schema != nil {
				err := lambda.Arguments[i].Schema.Validate(node)
				if err != nil {
					return nil, core.NewEvaluationErrorWithParent(node, "object does not satisfy type", err)
				}
			}

			newEnv.Vars[lambda.Arguments[i].Name] = node
		}

		return e.Eval(lambda.Body, newEnv, mode)

	case core.KindString:
		op, ok := car.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(car, fmt.Sprintf("invalid car value: %T", car.Value))
		}

		split := strings.Split(op, ".")
		if len(split) > 1 { // lib operator
			fn, ok := lib.GetOperator(split[0], split[1])
			if !ok {
				return nil, core.NewEvaluationError(car, fmt.Sprintf("unknown function name: %s", op))
			}

			// Call the operator function with the arguments
			return fn(cdr, env.CreateChild(), mode, e)

		} else { // built-in operator
			fn, ok := operators[op]
			if !ok {
				return nil, core.NewEvaluationError(car, fmt.Sprintf("unknown function name: %s", op))
			}

			// Call the operator function with the arguments
			return fn(cdr, env.CreateChild(), mode, e)
		}

	default:
		return nil, core.NewEvaluationError(car, fmt.Sprintf("cannot apply type %s", car.Kind))
	}
}

// operators is a map of operator names to their implementations
var operators = make(map[string]core.YispOperator)

// init initializes the operators map
func init() {
	// basic arithmetic operators
	operators["+"] = opAdd
	operators["add"] = opAdd
	operators["-"] = opSubtract
	operators["sub"] = opSubtract
	operators["*"] = opMultiply
	operators["mul"] = opMultiply
	operators["/"] = opDivide
	operators["div"] = opDivide

	// comparison operators
	operators["=="] = opEqual
	operators["eq"] = opEqual
	operators["!="] = opNotEqual
	operators["neq"] = opNotEqual
	operators["<"] = opLessThan
	operators["lt"] = opLessThan
	operators["<="] = opLessThanOrEqual
	operators["lte"] = opLessThanOrEqual
	operators[">"] = opGreaterThan
	operators["gt"] = opGreaterThan
	operators[">="] = opGreaterThanOrEqual
	operators["gte"] = opGreaterThanOrEqual

	// null coalesce operators
	operators["??"] = opNullCoalesce
	operators["default"] = opNullCoalesce

	// logical operators
	operators["&&"] = opAnd
	operators["and"] = opAnd
	operators["||"] = opOr
	operators["or"] = opOr
	operators["!"] = opNot
	operators["not"] = opNot

	// special operators
	operators["include"] = opInclude
	operators["progn"] = opProgn
	operators["pipeline"] = opPipeline
	operators["schema"] = opSchema
}

// opAdd adds numbers
func opAdd(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	sum := 0
	for _, node := range cdr {
		num, ok := node.Value.(int)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for +: %T", node))
		}
		sum += num
	}
	return &core.YispNode{
		Kind:  core.KindInt,
		Value: sum,
	}, nil
}

// opSubtract subtracts numbers
func opSubtract(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) == 0 {
		return &core.YispNode{
			Kind:  core.KindInt,
			Value: 0,
		}, nil
	}
	firstNode := cdr[0]
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, core.NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for -: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		val, ok := node.Value.(int)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for -: %T", node))
		}
		baseNum -= val
	}
	return &core.YispNode{
		Kind:  core.KindInt,
		Value: baseNum,
	}, nil
}

// opMultiply multiplies numbers
func opMultiply(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	product := 1
	for _, node := range cdr {
		num, ok := node.Value.(int)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for *: %T", node))
		}
		product *= num
	}
	return &core.YispNode{
		Kind:  core.KindInt,
		Value: product,
	}, nil
}

// opDivide divides numbers
func opDivide(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) == 0 {
		return &core.YispNode{
			Kind:  core.KindInt,
			Value: 0,
		}, nil
	}
	firstNode := cdr[0]
	baseNum, ok := firstNode.Value.(int)
	if !ok {
		return nil, core.NewEvaluationError(firstNode, fmt.Sprintf("invalid argument type for /: %T", firstNode))
	}
	for _, node := range cdr[1:] {
		val, ok := node.Value.(int)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid argument type for /: %T", node))
		}
		if val == 0 {
			return nil, core.NewEvaluationError(node, "division by zero")
		}
		baseNum /= val
	}
	return &core.YispNode{
		Kind:  core.KindInt,
		Value: baseNum,
	}, nil
}

// opEqual checks if two values are equal
func opEqual(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	return compareValues(cdr, "==", true)
}

// opNotEqual checks if two values are not equal
func opNotEqual(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	return compareValues(cdr, "!=", false)
}

// opLessThan checks if the first number is less than the second
func opLessThan(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	return compareNumbers(cdr, "<", func(a, b float64) bool { return a < b })
}

// opLessThanOrEqual checks if the first number is less than or equal to the second
func opLessThanOrEqual(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	return compareNumbers(cdr, "<=", func(a, b float64) bool { return a <= b })
}

// opGreaterThan checks if the first number is greater than the second
func opGreaterThan(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	return compareNumbers(cdr, ">", func(a, b float64) bool { return a > b })
}

// opGreaterThanOrEqual checks if the first number is greater than or equal to the second
func opGreaterThanOrEqual(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	return compareNumbers(cdr, ">=", func(a, b float64) bool { return a >= b })
}

func opNullCoalesce(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) == 0 {
		return &core.YispNode{
			Kind:  core.KindNull,
			Value: nil,
		}, nil
	}

	for _, node := range cdr {
		if node.Kind != core.KindNull {
			return node, nil
		}
	}

	return &core.YispNode{
		Kind:  core.KindNull,
		Value: nil,
	}, nil
}

// opAnd implements logical AND operation
func opAnd(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) == 0 {
		return &core.YispNode{
			Kind:  core.KindBool,
			Value: true,
		}, nil
	}

	for _, node := range cdr {
		truthy, err := core.IsTruthy(node)
		if err != nil {
			return nil, err
		}

		if !truthy {
			return &core.YispNode{
				Kind:  core.KindBool,
				Value: false,
			}, nil
		}
	}

	return &core.YispNode{
		Kind:  core.KindBool,
		Value: true,
	}, nil
}

// opOr implements logical OR operation
func opOr(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) == 0 {
		return &core.YispNode{
			Kind:  core.KindBool,
			Value: false,
		}, nil
	}

	for _, node := range cdr {
		truthy, err := core.IsTruthy(node)
		if err != nil {
			return nil, err
		}

		if truthy {
			return &core.YispNode{
				Kind:  core.KindBool,
				Value: true,
			}, nil
		}
	}

	return &core.YispNode{
		Kind:  core.KindBool,
		Value: false,
	}, nil
}

// opNot implements logical NOT operation
func opNot(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("not requires 1 argument, got %d", len(cdr)))
	}

	node := cdr[0]
	truthy, err := core.IsTruthy(node)
	if err != nil {
		return nil, err
	}

	return &core.YispNode{
		Kind:  core.KindBool,
		Value: !truthy,
	}, nil
}

// opInclude includes files
func opInclude(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	results := make([]any, 0)
	for _, node := range cdr {
		relpath, ok := node.Value.(string)
		if !ok {
			return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid path type: %T", node.Value))
		}

		var err error
		evaluated, err := core.CallEngineByPath(relpath, node.Attr.File(), core.NewEnv(), e)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(node, "failed to include file", err)
		}

		if evaluated.Kind == core.KindArray {
			arr, ok := evaluated.Value.([]any)
			if !ok {
				return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid array value: %T", evaluated.Value))
			}

			for _, item := range arr {
				itemNode, ok := item.(*core.YispNode)
				if !ok {
					return nil, core.NewEvaluationError(node, fmt.Sprintf("invalid item type: %T", item))
				}
				itemNode.Tag = "!quote"
				results = append(results, itemNode)
			}
		} else {
			evaluated.Tag = "!quote"
			results = append(results, evaluated)
		}
	}

	return &core.YispNode{
		Kind:           core.KindArray,
		Value:          results,
		IsDocumentRoot: true,
	}, nil
}

func opProgn(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	return cdr[len(cdr)-1], nil
}

func opPipeline(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {

	value := cdr[0]

	for _, fn := range cdr[1:] {
		var err error
		value, err = e.Apply(fn, []*core.YispNode{value}, env, mode)
		if err != nil {
			return nil, core.NewEvaluationErrorWithParent(fn, "failed to evaluate pipeline", err)
		}
	}

	return value, nil
}

func opSchema(cdr []*core.YispNode, env *core.Env, mode core.EvalMode, e core.Engine) (*core.YispNode, error) {
	if len(cdr) != 1 {
		return nil, core.NewEvaluationError(nil, fmt.Sprintf("sha256 requires 1 argument, got %d", len(cdr)))
	}
	node := cdr[0]
	if node.Kind != core.KindMap {
		return nil, core.NewEvaluationError(node, fmt.Sprintf("schema requires a map argument, got %v", node.Kind))
	}

	rendered, err := ToNative(node)
	if err != nil {
		return nil, core.NewEvaluationErrorWithParent(node, "failed to render schema", err)
	}
	schemaBytes, err := json.Marshal(rendered)
	if err != nil {
		return nil, core.NewEvaluationErrorWithParent(node, "failed to marshal schema", err)
	}

	var schema core.Schema
	err = json.Unmarshal(schemaBytes, &schema)
	if err != nil {
		return nil, core.NewEvaluationErrorWithParent(node, "failed to unmarshal schema", err)
	}

	return &core.YispNode{
		Kind:  core.KindType,
		Value: &schema,
		Attr:  node.Attr,
	}, nil

}
