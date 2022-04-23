package main

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

type TokenKind string

const (
	KindIdentifier TokenKind = "identifier" // `abc`, `a12`, `a_b_1`
	KindNumber     TokenKind = "number"     // `123`, `1.12`, `-12`, `1_2_3`
	KindOperator   TokenKind = "operator"   // `+`, `-`, `^`, `(`
)

var constants = map[string]decimal.Decimal{
	"Pi": decimal.NewFromFloat(math.Pi),
	"e":  decimal.NewFromFloat(math.E),
}

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
	OpMod
	OpRoot
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
	"%":  OpMod,
	"@":  OpRoot,
}

var opsToSymbolText = map[Operator]string{
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
	OpMod:         "%",
	OpRoot:        "@",
}

var opsToText = map[Operator]string{
	OpNoOp:        "no-op",
	OpPlus:        "plus",
	OpMinus:       "minus",
	OpUnaryMinus:  "unary minus",
	OpMultiply:    "multiply",
	OpDivide:      "divide",
	OpPower:       "power",
	OpOpenParent:  "open parent",
	OpCloseParent: "close parent",
	OpInc:         "increment",
	OpDec:         "decrement",
	OpMod:         "mod",
	OpRoot:        "root",
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
	OpMod:         TypeBinary,
	OpRoot:        TypeBinary,
}

func opPrecedence(op Operator) int {
	switch op {
	case OpPlus, OpMinus:
		return 1
	case OpMultiply, OpDivide, OpMod:
		return 2
	case OpPower, OpRoot:
		return 3
	case OpUnaryMinus, OpInc, OpDec:
		return 4
	default:
		return 0
	}
}

func numberOrVariable(token Token, variables map[Token]decimal.Decimal) (decimal.Decimal, bool) {
	switch token.kind {
	case KindNumber:
		return token.number, true
	case KindIdentifier:
		number, ok := variables[token]
		return number, ok
	default:
		return decimal.Decimal{}, false
	}
}

func applyUnaryOp(v, op Token, variables map[Token]decimal.Decimal) (Token, bool) {
	number, ok := numberOrVariable(v, variables)
	if !ok {
		return v, false
	}

	var result decimal.Decimal
	switch op.op {
	case OpUnaryMinus:
		result = number.Neg()
	case OpInc:
		result = number.Add(one)
	case OpDec:
		result = number.Sub(one)
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
		text: fmt.Sprintf("%s %s", opsToSymbolText[op.op], v.text),
		loc: Location{
			start: op.loc.start,
			end:   v.loc.end,
		},
		number: result,
		op:     0,
	}, true
}

func applyBinaryOp(v1, v2, op Token, variables map[Token]decimal.Decimal) (Token, bool) {
	number1, ok := numberOrVariable(v1, variables)
	if !ok {
		return v1, false
	}

	number2, ok := numberOrVariable(v2, variables)
	if !ok {
		return v2, false
	}

	var result decimal.Decimal
	switch op.op {
	case OpPlus:
		result = number1.Add(number2)
	case OpMinus:
		result = number1.Sub(number2)
	case OpMultiply:
		result = number1.Mul(number2)
	case OpDivide:
		if number2.IsZero() {
			return v2, false
		}
		result = number1.Div(number2)
	case OpPower:
		if !number2.IsInteger() {
			return v2, false
		}
		result = number1.Pow(number2)
	case OpMod:
		if !number1.IsInteger() {
			return v1, false
		}
		if !number2.IsInteger() {
			return v2, false
		}
		result = number1.Mod(number2)
	case OpRoot:
		if !number2.IsInteger() {
			return v2, false
		}
		result = DecimalRoot(number1, number2)
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
		text: fmt.Sprintf("%s %s %s", v1.text, opsToSymbolText[op.op], v2.text),
		loc: Location{
			start: v1.loc.start,
			end:   v2.loc.end,
		},
		number: result,
		op:     0,
	}, true
}
