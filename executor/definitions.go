package executor

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"

	"github.com/mymmrac/mm/utils"
)

var constants = map[string]decimal.Decimal{
	"Pi": decimal.NewFromFloat(math.Pi),
	"e":  decimal.NewFromFloat(math.E),
}

type Operator int

const (
	OpNoOp Operator = iota
	OpOpenParent
	OpCloseParent
	OpPlus
	OpMinus
	OpUnaryMinus
	OpMultiply
	OpDivide
	OpPower
	OpInc
	OpDec
	OpMod
	OpRoot
	OpRound
	OpFloor
	OpCeil
	OpAbs
)

var textToOp = map[string]Operator{
	"(":      OpOpenParent,
	")":      OpCloseParent,
	"+":      OpPlus,
	"-":      OpMinus,
	"*":      OpMultiply,
	"/":      OpDivide,
	"^":      OpPower,
	"++":     OpInc,
	"--":     OpDec,
	"%":      OpMod,
	"@":      OpRoot,
	"!round": OpRound,
	"!floor": OpFloor,
	"!ceil":  OpCeil,
	"!abs":   OpAbs,
}

var opToSymbolText = map[Operator]string{
	OpNoOp: "no-op",

	OpOpenParent:  "(",
	OpCloseParent: ")",
	OpPlus:        "+",
	OpMinus:       "-",
	OpUnaryMinus:  "-",
	OpMultiply:    "*",
	OpDivide:      "/",
	OpPower:       "^",
	OpInc:         "++",
	OpDec:         "--",
	OpMod:         "%",
	OpRoot:        "@",
	OpRound:       "!round",
	OpFloor:       "!floor",
	OpCeil:        "!ceil",
	OpAbs:         "!abs",
}

var opToText = map[Operator]string{
	OpNoOp:        "no-op",
	OpOpenParent:  "open parent",
	OpCloseParent: "close parent",
	OpPlus:        "plus",
	OpMinus:       "minus",
	OpUnaryMinus:  "unary minus",
	OpMultiply:    "multiply",
	OpDivide:      "divide",
	OpPower:       "power",
	OpInc:         "increment",
	OpDec:         "decrement",
	OpMod:         "mod",
	OpRoot:        "root",
	OpRound:       "round",
	OpFloor:       "floor",
	OpCeil:        "ceil",
	OpAbs:         "abs",
}

type OpType string

const (
	TypeNoOp   OpType = "no-op"
	TypeUnary  OpType = "unary"
	TypeBinary OpType = "binary"
)

var opTypes = map[Operator]OpType{
	OpOpenParent:  TypeNoOp,
	OpCloseParent: TypeNoOp,
	OpPlus:        TypeBinary,
	OpMinus:       TypeBinary,
	OpMultiply:    TypeBinary,
	OpDivide:      TypeBinary,
	OpPower:       TypeBinary,
	OpUnaryMinus:  TypeUnary,
	OpInc:         TypeUnary,
	OpDec:         TypeUnary,
	OpMod:         TypeBinary,
	OpRoot:        TypeBinary,
	OpRound:       TypeUnary,
	OpFloor:       TypeUnary,
	OpCeil:        TypeUnary,
	OpAbs:         TypeUnary,
}

func opPrecedence(op Operator) int {
	switch op {
	case OpOpenParent, OpCloseParent:
		return 0
	case OpPlus, OpMinus:
		return 1
	case OpMultiply, OpDivide, OpMod:
		return 2
	case OpPower, OpRoot:
		return 3
	case OpInc, OpDec, OpUnaryMinus, OpRound, OpFloor, OpCeil, OpAbs:
		return 4
	default:
		utils.Assert(false, "unreachable")
		return 0
	}
}

func numberOrVariable(token Token, variables Vars) (decimal.Decimal, bool) {
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

func applyUnaryOp(v, op Token, variables Vars) (Token, bool) {
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
	case OpRound:
		result = number.Round(0)
	case OpFloor:
		result = number.Floor()
	case OpCeil:
		result = number.Ceil()
	case OpAbs:
		result = number.Abs()
	default:
		return Token{
			loc: Location{
				Start: op.loc.Start,
				End:   v.loc.End,
			},
		}, false
	}

	return Token{
		kind: KindNumber,
		text: fmt.Sprintf("%s %s", opToSymbolText[op.op], v.text),
		loc: Location{
			Start: op.loc.Start,
			End:   v.loc.End,
		},
		number: result,
		op:     0,
	}, true
}

func applyBinaryOp(v1, v2, op Token, variables Vars) (Token, bool) {
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
		if number1.IsNegative() {
			return v1, false
		}
		if !number2.IsInteger() {
			return v2, false
		}
		result = DecimalRoot(number1, number2)
	default:
		return Token{
			loc: Location{
				Start: v1.loc.Start,
				End:   v2.loc.End,
			},
		}, false
	}

	return Token{
		kind: KindNumber,
		text: fmt.Sprintf("%s %s %s", v1.text, opToSymbolText[op.op], v2.text),
		loc: Location{
			Start: v1.loc.Start,
			End:   v2.loc.End,
		},
		number: result,
		op:     0,
	}, true
}
