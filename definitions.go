package main

import "math"

const (
	OpNone Operator = iota
	OpPlus
	OpMinus
	OpMultiply
	OpDivide
	OpPower
	OpOpenParent
	OpClosedParent

	OpsLast
)

var textToOps = map[string]Operator{
	"+": OpPlus,
	"-": OpMinus,
	"*": OpMultiply,
	"/": OpDivide,
	"^": OpPower,
	"(": OpOpenParent,
	")": OpClosedParent,
}
var opsToText map[Operator]string

func opPrecedence(op Operator) int {
	switch op {
	case OpPlus, OpMinus:
		return 1
	case OpMultiply, OpDivide:
		return 2
	case OpPower:
		return 3
	default:
		return 0
	}
}

func applyBinaryOp(a, b float64, op Operator) (float64, bool) {
	switch op {
	case OpPlus:
		return a + b, true
	case OpMinus:
		return a - b, true
	case OpMultiply:
		return a * b, true
	case OpDivide:
		return a / b, true
	case OpPower:
		return math.Pow(a, b), true
	default:
		return 0, false
	}
}
