package main

import (
	"fmt"
	"math"
)

type TokenKind string

const (
	KindIdentifier TokenKind = "identifier" // `abc`, `a12`, `a_b_1`
	KindNumber     TokenKind = "number"     // `123`, `1.12`, `-12`, `1_2_3`
	KindOperator   TokenKind = "operator"   // `+`, `-`, `^`, `(`
)

type Operator int

const (
	OpNoOp Operator = iota
	OpPlus
	OpMinus
	OpMultiply
	OpDivide
	OpPower
	OpOpenParent
	OpClosedParent
	OpUnaryMinus

	OpsLast
)

var textToOps = map[string]Operator{
	"+":  OpPlus,
	"-":  OpMinus,
	"*":  OpMultiply,
	"/":  OpDivide,
	"^":  OpPower,
	"(":  OpOpenParent,
	")":  OpClosedParent,
	"--": OpUnaryMinus,
}
var opsToText map[Operator]string

type OpType string

const (
	TypeNoOp   OpType = "no-op"
	TypeUnary  OpType = "unary"
	TypeBinary OpType = "binary"
)

var opsTypes = map[Operator]OpType{
	OpPlus:         TypeBinary,
	OpMinus:        TypeBinary,
	OpMultiply:     TypeBinary,
	OpDivide:       TypeBinary,
	OpPower:        TypeBinary,
	OpOpenParent:   TypeNoOp,
	OpClosedParent: TypeNoOp,
	OpUnaryMinus:   TypeUnary,
}

func opPrecedence(op Operator) int {
	switch op {
	case OpPlus, OpMinus:
		return 1
	case OpMultiply, OpDivide:
		return 2
	case OpPower:
		return 3
	case OpUnaryMinus:
		return 4
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

func applyUnaryOp(a Token, op Operator) (Token, bool) {
	if a.kind != KindNumber {
		return Token{
			loc: Location{
				start: a.loc.start,
				end:   a.loc.end, // TODO: Op as Token
			},
		}, false
	}

	var result float64
	switch op {
	case OpUnaryMinus:
		result = -a.number
	default:
		return Token{
			loc: Location{
				start: a.loc.start,
				end:   a.loc.end,
			},
		}, false
	}

	return Token{
		kind: KindNumber,
		text: fmt.Sprintf("%s %s", opsToText[op], a.text),
		loc: Location{
			start: a.loc.start,
			end:   a.loc.end,
		},
		number: result,
		op:     0,
	}, true
}
