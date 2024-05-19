package executor

import (
	"fmt"
	"slices"

	"github.com/shopspring/decimal"

	"github.com/mymmrac/mm/utils"
)

const defaultPrecision = 32

type Operator struct {
	text       string
	name       string
	precedence uint
	arity      uint
	apply      func(stack *utils.Stack[decimal.Decimal]) error
}

var (
	opOpenParenthesis  = Operator{text: "(", name: "open parenthesis"}
	opCloseParenthesis = Operator{text: ")", name: "close parenthesis"}
	opComma            = Operator{text: ",", name: "comma"}
)

var knownOperators = []Operator{
	opOpenParenthesis,
	opCloseParenthesis,
	opComma,

	{
		text:       "+",
		name:       "addition",
		precedence: 1,
		arity:      2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Add(v2), nil
		}),
	},
	{
		text:       "+",
		name:       "unary plus",
		precedence: 4,
		arity:      1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1, nil
		}),
	},
	{
		text:       "-",
		name:       "subtraction",
		precedence: 1,
		arity:      2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Sub(v2), nil
		}),
	},
	{
		text:       "-",
		name:       "unary minus",
		precedence: 4,
		arity:      1,
		apply: applyUnaryOp(func(v1 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Neg(), nil
		}),
	},
	{
		text:       "*",
		name:       "multiplication",
		precedence: 2,
		arity:      2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Mul(v2), nil
		}),
	},
	{
		text:       "/",
		name:       "division",
		precedence: 2,
		arity:      2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			if v2.IsZero() {
				return decimal.Zero, fmt.Errorf("division by zero")
			}
			return v1.DivRound(v2, defaultPrecision), nil
		}),
	},
	{
		text:       "//",
		name:       "floor division",
		precedence: 2,
		arity:      2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			if v2.IsZero() {
				return decimal.Zero, fmt.Errorf("division by zero")
			}
			return v1.DivRound(v2, 0), nil
		}),
	},
	{
		text:       "^",
		name:       "power",
		precedence: 3,
		arity:      2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			switch {
			case v1.IsZero() && v2.IsZero():
				return decimal.Zero, fmt.Errorf("undefined value (0 ^ 0)")
			case v1.IsZero() && v2.IsNegative():
				return decimal.Zero, fmt.Errorf("infinity")
			case v1.IsNegative() && !v2.IsInteger():
				return decimal.Zero, fmt.Errorf("imaginary value")
			}
			return v1.PowWithPrecision(v2, defaultPrecision)
		}),
	},
	{
		text:       "%",
		name:       "modulo",
		precedence: 2,
		arity:      2,
		apply: applyBinaryOp(func(v1, v2 decimal.Decimal) (decimal.Decimal, error) {
			return v1.Mod(v2), nil
		}),
	},
}

var knownUniqueOperators []string

func init() {
	// TODO: Validate that [text] + [arity] is unique
	// TODO: Validate that [arity] is 1 or 2 if not `(`, `)`, `,`

	for _, operator := range knownOperators {
		knownUniqueOperators = append(knownUniqueOperators, operator.text)
	}
	slices.SortFunc(knownUniqueOperators, func(a, b string) int {
		return len(b) - len(a)
	})
	knownUniqueOperators = slices.Compact(knownUniqueOperators)
}

func applyUnaryOp(
	apply func(v1 decimal.Decimal) (decimal.Decimal, error),
) func(stack *utils.Stack[decimal.Decimal]) error {
	return func(stack *utils.Stack[decimal.Decimal]) error {
		v1 := stack.Pop()

		result, err := apply(v1)
		if err != nil {
			return err
		}
		stack.Push(result)

		return nil
	}
}

func applyBinaryOp(
	apply func(v1, v2 decimal.Decimal) (decimal.Decimal, error),
) func(stack *utils.Stack[decimal.Decimal]) error {
	return func(stack *utils.Stack[decimal.Decimal]) error {
		v2 := stack.Pop()
		v1 := stack.Pop()

		result, err := apply(v1, v2)
		if err != nil {
			return err
		}
		stack.Push(result)

		return nil
	}
}
