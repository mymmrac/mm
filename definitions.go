package main

import (
	"fmt"
	"math"
)

const (
	OpNoOp Operator = iota
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

func applyBinaryOp(a, b Token, op Operator) (Token, bool) {
	if a.kind != KindNumber || b.kind != KindNumber {
		return Token{
			loc: Location{
				start: a.loc.start,
				end:   b.loc.end,
			},
		}, false
	}

	var result float64
	switch op {
	case OpPlus:
		result = a.number + b.number
	case OpMinus:
		result = a.number - b.number
	case OpMultiply:
		result = a.number * b.number
	case OpDivide:
		result = a.number / b.number
	case OpPower:
		result = math.Pow(a.number, b.number)
	default:
		return Token{
			loc: Location{
				start: a.loc.start,
				end:   b.loc.end,
			},
		}, false
	}

	return Token{
		kind: KindNumber,
		text: fmt.Sprintf("%s %s %s", a.text, opsToText[op], b.text),
		loc: Location{
			start: a.loc.start,
			end:   b.loc.end,
		},
		number: result,
		op:     0,
	}, true
}
