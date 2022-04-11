package main

import (
	"fmt"

	"github.com/shopspring/decimal"
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
	OpUnaryMinus
	OpMultiply
	OpDivide
	OpPower
	OpOpenParent
	OpCloseParent
	OpInc
	OpDec
)

var textToOps = map[string]Operator{
	"+":  OpPlus,
	"-":  OpMinus,
	"*":  OpMultiply,
	"/":  OpDivide,
	"^":  OpPower,
	"(":  OpOpenParent,
	")":  OpCloseParent,
	"++": OpInc,
	"--": OpDec,
}

var opsToText = map[Operator]string{
	OpNoOp: "no-op",

	OpPlus:        "+",
	OpMinus:       "-",
	OpUnaryMinus:  "-",
	OpMultiply:    "*",
	OpDivide:      "/",
	OpPower:       "^",
	OpOpenParent:  "(",
	OpCloseParent: ")",
	OpInc:         "++",
	OpDec:         "--",
}

type OpType string

const (
	TypeNoOp   OpType = "no-op"
	TypeUnary  OpType = "unary"
	TypeBinary OpType = "binary"
)

var opsTypes = map[Operator]OpType{
	OpPlus:        TypeBinary,
	OpMinus:       TypeBinary,
	OpMultiply:    TypeBinary,
	OpDivide:      TypeBinary,
	OpPower:       TypeBinary,
	OpOpenParent:  TypeNoOp,
	OpCloseParent: TypeNoOp,
	OpUnaryMinus:  TypeUnary,
	OpInc:         TypeUnary,
	OpDec:         TypeUnary,
}

func opPrecedence(op Operator) int {
	switch op {
	case OpPlus, OpMinus:
		return 1
	case OpMultiply, OpDivide:
		return 2
	case OpPower:
		return 3
	case OpUnaryMinus, OpInc, OpDec:
		return 4
	default:
		return 0
	}
}

func applyUnaryOp(v, op Token) (Token, bool) {
	if v.kind != KindNumber {
		return Token{
			loc: Location{
				start: op.loc.start,
				end:   v.loc.end,
			},
		}, false
	}

	one := decimal.NewFromInt(1)

	var result decimal.Decimal
	switch op.op {
	case OpUnaryMinus:
		result = v.number.Neg()
	case OpInc:
		result = v.number.Add(one)
	case OpDec:
		result = v.number.Sub(one)
	default:
		return Token{
			loc: Location{
				start: op.loc.start,
				end:   v.loc.end,
			},
		}, false
	}

	return Token{
		kind: KindNumber,
		text: fmt.Sprintf("%s %s", opsToText[op.op], v.text),
		loc: Location{
			start: op.loc.start,
			end:   v.loc.end,
		},
		number: result,
		op:     0,
	}, true
}

func applyBinaryOp(v1, v2, op Token) (Token, bool) {
	if v1.kind != KindNumber || v2.kind != KindNumber {
		return Token{
			loc: Location{
				start: v1.loc.start,
				end:   v2.loc.end,
			},
		}, false
	}

	var result decimal.Decimal
	switch op.op {
	case OpPlus:
		result = v1.number.Add(v2.number)
	case OpMinus:
		result = v1.number.Sub(v2.number)
	case OpMultiply:
		result = v1.number.Mul(v2.number)
	case OpDivide:
		result = v1.number.Div(v2.number)
	case OpPower:
		result = v1.number.Pow(v2.number)
	default:
		return Token{
			loc: Location{
				start: v1.loc.start,
				end:   v2.loc.end,
			},
		}, false
	}

	return Token{
		kind: KindNumber,
		text: fmt.Sprintf("%s %s %s", v1.text, opsToText[op.op], v2.text),
		loc: Location{
			start: v1.loc.start,
			end:   v2.loc.end,
		},
		number: result,
		op:     0,
	}, true
}
