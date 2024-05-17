package executor

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/shopspring/decimal"

	"github.com/mymmrac/mm/utils"
)

var (
	constPi, _ = decimal.NewFromString("3.14159265358979323846264338327950288419716939937510582097494459")
	constE, _  = decimal.NewFromString("2.71828182845904523536028747135266249775724709369995957496696763")
)

type Identifier struct {
	text     string
	name     string
	variable bool
	arity    uint
	apply    func(stack *utils.Stack[decimal.Decimal]) error
}

var knownIdentifiers = []Identifier{
	{
		text:     "Pi",
		name:     "number Pi",
		variable: true,
		apply:    applyConstantIdent(constPi),
	},
	{
		text:     "e",
		name:     "number e",
		variable: true,
		apply:    applyConstantIdent(constE),
	},
	{
		text:  "sqrt",
		name:  "square root",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			if v1.IsNegative() {
				return decimal.Zero, fmt.Errorf("square root of negative number")
			}
			return v1.PowWithPrecision(decimal.NewFromFloat(0.5), 32)
		}),
	},
	{
		text:  "abs",
		name:  "absolute value",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Abs(), nil
		}),
	},
	{
		text:  "round",
		name:  "round",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Round(0), nil
		}),
	},
	{
		text:  "round",
		name:  "round",
		arity: 2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			if !v2.IsInteger() {
				return decimal.Zero, fmt.Errorf("the second argument must be an integer")
			}
			return v1.Round(int32(v2.IntPart())), nil
		}),
	},
	{
		text:  "roundUp",
		name:  "round up",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.RoundUp(0), nil
		}),
	},
	{
		text:  "roundUp",
		name:  "round up",
		arity: 2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			if !v2.IsInteger() {
				return decimal.Zero, fmt.Errorf("the second argument must be an integer")
			}
			return v1.RoundUp(int32(v2.IntPart())), nil
		}),
	},
	{
		text:  "floor",
		name:  "floor",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Floor(), nil
		}),
	},
	{
		text:  "ceil",
		name:  "ceil",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Ceil(), nil
		}),
	},
	{
		text:  "sin",
		name:  "sine",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Sin(), nil
		}),
	},
	{
		text:  "cos",
		name:  "cosine",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Cos(), nil
		}),
	},
	{
		text:  "tan",
		name:  "tangent",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Tan(), nil
		}),
	},
	{
		text:  "rad",
		name:  "radian",
		arity: 1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			if v1.IsZero() {
				return decimal.Zero, nil
			}
			return constPi.Mul(v1).DivRound(decimal.NewFromInt(180), defaultPrecision), nil
		}),
	},
	{
		text:  "min",
		name:  "minimum",
		arity: 2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			if v1.Cmp(v2) < 0 {
				return v1, nil
			}
			return v2, nil
		}),
	},
	{
		text:  "max",
		name:  "maximum",
		arity: 2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			if v2.Cmp(v1) < 0 {
				return v1, nil
			}
			return v2, nil
		}),
	},
	{
		text:  "rand",
		name:  "random number",
		arity: 0,
		apply: applyNullaryIdent(func() (decimal.Decimal, error) {
			return decimal.NewFromFloat(rand.Float64()), nil
		}),
	},
}

var knownUniqueIdentifiers []string

func init() {
	// TODO: Validate that [text] + [arity] + [variable] is unique (also with operators)
	// TODO: Validate that [text] doesn't start with digit
	// TODO: Validate that if [variable] then [arity] == 0

	for _, identifier := range knownIdentifiers {
		knownUniqueIdentifiers = append(knownUniqueIdentifiers, identifier.text)
	}
	slices.SortFunc(knownUniqueIdentifiers, func(a, b string) int {
		return len(b) - len(a)
	})
	knownUniqueIdentifiers = slices.Compact(knownUniqueIdentifiers)
}

func applyConstantIdent(constant decimal.Decimal) func(stack *utils.Stack[decimal.Decimal]) error {
	return func(stack *utils.Stack[decimal.Decimal]) error {
		stack.Push(constant)
		return nil
	}
}

func applyNullaryIdent(apply func() (decimal.Decimal, error)) func(stack *utils.Stack[decimal.Decimal]) error {
	return func(stack *utils.Stack[decimal.Decimal]) error {
		result, err := apply()
		if err != nil {
			return err
		}
		stack.Push(result)
		return nil
	}
}
