package core

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ExprType is the data type of an evaluated expression value.
type ExprType string

const (
	ExprTypeNull    ExprType = "null"
	ExprTypeBoolean ExprType = "boolean"
	ExprTypeNumber  ExprType = "number"
	ExprTypeString  ExprType = "string"
	ExprTypeArray   ExprType = "array"
	ExprTypeObject  ExprType = "object"
)

// ExprValue is a dynamically typed value produced by expression evaluation.
type ExprValue struct {
	Type    ExprType
	BoolVal bool
	NumVal  float64
	StrVal  string
	ArrVal  []ExprValue
	ObjVal  map[string]ExprValue
}

var ExprNull = ExprValue{Type: ExprTypeNull}
var ExprTrue = ExprValue{Type: ExprTypeBoolean, BoolVal: true}
var ExprFalse = ExprValue{Type: ExprTypeBoolean, BoolVal: false}

// ExprNum creates a numeric ExprValue.
func ExprNum(n float64) ExprValue { return ExprValue{Type: ExprTypeNumber, NumVal: n} }

// ExprStr creates a string ExprValue.
func ExprStr(s string) ExprValue { return ExprValue{Type: ExprTypeString, StrVal: s} }

// ExprBool creates a boolean ExprValue.
func ExprBool(b bool) ExprValue {
	if b {
		return ExprTrue
	}
	return ExprFalse
}

// Truthy evaluates whether an ExprValue is considered logically true.
func (v ExprValue) Truthy() bool {
	switch v.Type {
	case ExprTypeNull:
		return false
	case ExprTypeBoolean:
		return v.BoolVal
	case ExprTypeNumber:
		return v.NumVal != 0
	case ExprTypeString:
		return v.StrVal != ""
	case ExprTypeArray:
		return len(v.ArrVal) > 0
	case ExprTypeObject:
		return len(v.ObjVal) > 0
	}
	return false
}

// String renders the value as a human-readable string.
func (v ExprValue) String() string {
	switch v.Type {
	case ExprTypeNull:
		return "null"
	case ExprTypeBoolean:
		return strconv.FormatBool(v.BoolVal)
	case ExprTypeNumber:
		return strconv.FormatFloat(v.NumVal, 'f', -1, 64)
	case ExprTypeString:
		return v.StrVal
	default:
		return fmt.Sprintf("<%s>", v.Type)
	}
}

// EvalContext provides variable bindings for expression evaluation.
type EvalContext struct {
	vars map[string]ExprValue
}

// NewEvalContext creates an empty evaluation context.
func NewEvalContext() *EvalContext {
	return &EvalContext{vars: make(map[string]ExprValue)}
}

// Set binds a variable in the evaluation context.
func (e *EvalContext) Set(name string, val ExprValue) *EvalContext {
	e.vars[name] = val
	return e
}

// Get retrieves a variable from the evaluation context.
func (e *EvalContext) Get(name string) (ExprValue, bool) {
	v, ok := e.vars[name]
	return v, ok
}

// ExprEvaluator evaluates workflow step condition expressions from JSON-serialized AST nodes.
type ExprEvaluator struct {
	ctx *EvalContext
}

// NewExprEvaluator creates an evaluator with a given context.
func NewExprEvaluator(ctx *EvalContext) *ExprEvaluator {
	return &ExprEvaluator{ctx: ctx}
}

// EvalJSON evaluates a JSON-encoded expression definition against the context.
func (e *ExprEvaluator) EvalJSON(raw json.RawMessage) (ExprValue, error) {
	var node map[string]interface{}
	if err := json.Unmarshal(raw, &node); err != nil {
		// Try literal value
		return e.evalLiteral(raw)
	}
	return e.evalNode(node)
}

func (e *ExprEvaluator) evalNode(node map[string]interface{}) (ExprValue, error) {
	op, _ := node["op"].(string)

	switch op {
	case "var":
		name, _ := node["name"].(string)
		if val, ok := e.ctx.Get(name); ok {
			return val, nil
		}
		return ExprNull, nil

	case "lit":
		raw, _ := node["value"]
		return exprFromInterface(raw)

	case "not":
		sub, err := e.evalSubExpr(node, "expr")
		if err != nil {
			return ExprNull, err
		}
		return ExprBool(!sub.Truthy()), nil

	case "and":
		left, err := e.evalSubExpr(node, "left")
		if err != nil {
			return ExprNull, err
		}
		if !left.Truthy() {
			return ExprFalse, nil
		}
		return e.evalSubExpr(node, "right")

	case "or":
		left, err := e.evalSubExpr(node, "left")
		if err != nil {
			return ExprNull, err
		}
		if left.Truthy() {
			return left, nil
		}
		return e.evalSubExpr(node, "right")

	case "eq":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		return ExprBool(l.String() == r.String()), nil

	case "neq":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		return ExprBool(l.String() != r.String()), nil

	case "lt":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		if l.Type == ExprTypeNumber && r.Type == ExprTypeNumber {
			return ExprBool(l.NumVal < r.NumVal), nil
		}
		return ExprBool(l.String() < r.String()), nil

	case "lte":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		if l.Type == ExprTypeNumber && r.Type == ExprTypeNumber {
			return ExprBool(l.NumVal <= r.NumVal), nil
		}
		return ExprBool(l.String() <= r.String()), nil

	case "gt":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		if l.Type == ExprTypeNumber && r.Type == ExprTypeNumber {
			return ExprBool(l.NumVal > r.NumVal), nil
		}
		return ExprBool(l.String() > r.String()), nil

	case "add":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		if l.Type == ExprTypeNumber && r.Type == ExprTypeNumber {
			return ExprNum(l.NumVal + r.NumVal), nil
		}
		return ExprStr(l.String() + r.String()), nil

	case "sub":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		return ExprNum(l.NumVal - r.NumVal), nil

	case "mul":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		return ExprNum(l.NumVal * r.NumVal), nil

	case "div":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		if r.NumVal == 0 {
			return ExprNull, fmt.Errorf("%w: division by zero", ErrValidationFailed)
		}
		return ExprNum(l.NumVal / r.NumVal), nil

	case "abs":
		sub, err := e.evalSubExpr(node, "expr")
		if err != nil {
			return ExprNull, err
		}
		return ExprNum(math.Abs(sub.NumVal)), nil

	case "contains":
		l, r, err := e.evalBinaryOperands(node)
		if err != nil {
			return ExprNull, err
		}
		return ExprBool(strings.Contains(l.String(), r.String())), nil

	case "if":
		cond, err := e.evalSubExpr(node, "cond")
		if err != nil {
			return ExprNull, err
		}
		if cond.Truthy() {
			return e.evalSubExpr(node, "then")
		}
		return e.evalSubExpr(node, "else")
	}

	return ExprNull, fmt.Errorf("unknown expression op: %q", op)
}

func (e *ExprEvaluator) evalSubExpr(node map[string]interface{}, key string) (ExprValue, error) {
	sub, ok := node[key]
	if !ok {
		return ExprNull, fmt.Errorf("missing key %q in expression node", key)
	}
	if subMap, ok := sub.(map[string]interface{}); ok {
		return e.evalNode(subMap)
	}
	return exprFromInterface(sub)
}

func (e *ExprEvaluator) evalBinaryOperands(node map[string]interface{}) (ExprValue, ExprValue, error) {
	l, err := e.evalSubExpr(node, "left")
	if err != nil {
		return ExprNull, ExprNull, fmt.Errorf("left operand: %w", err)
	}
	r, err := e.evalSubExpr(node, "right")
	if err != nil {
		return ExprNull, ExprNull, fmt.Errorf("right operand: %w", err)
	}
	return l, r, nil
}

func (e *ExprEvaluator) evalLiteral(raw json.RawMessage) (ExprValue, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return ExprStr(s), nil
	}
	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		return ExprNum(n), nil
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return ExprBool(b), nil
	}
	return ExprNull, nil
}

func exprFromInterface(v interface{}) (ExprValue, error) {
	switch val := v.(type) {
	case nil:
		return ExprNull, nil
	case bool:
		return ExprBool(val), nil
	case float64:
		return ExprNum(val), nil
	case string:
		return ExprStr(val), nil
	}
	return ExprNull, nil
}
