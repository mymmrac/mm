package main

import (
	"strconv"

	"github.com/mymmrac/mm/utils"
)

type Location struct {
	start int
	end   int
}

func (l Location) Size() int {
	return l.end - l.start
}

type ExprError struct {
	text string
	loc  Location
}

func NewExprErr(text string, loc Location) *ExprError {
	return &ExprError{
		text: text,
		loc:  loc,
	}
}

type Executor struct {
	lexer *Lexer
}

func NewExecutor() *Executor {
	return &Executor{
		lexer: NewLexer(),
	}
}

func (e *Executor) Execute(expr string) (string, *ExprError) {
	tokens, err := e.lexer.Tokenize(expr)
	if err != nil {
		return "", err
	}

	if err = e.typeCheck(tokens); err != nil {
		return "", err
	}

	res := e.evaluate(tokens)
	return strconv.FormatFloat(res, 'f', -1, 64), nil
}

func (e *Executor) typeCheck(tokens []Token) *ExprError {
	for i, token := range tokens {
		switch token.kind {
		case KindIdentifier:
		case KindNumber:
			n, err := strconv.ParseFloat(token.text, 64)
			if err != nil {
				return NewExprErr("parsing number: "+err.Error(), token.loc)
			}
			tokens[i].number = n
		case KindOperator:
			op, ok := textToOps[token.text]
			if !ok {
				return NewExprErr("unknown operator: "+token.text, token.loc)
			}
			tokens[i].op = op
		default:
			utils.Assert(false, "unreachable")
		}
	}

	return nil
}

func (e *Executor) precedence(op Operator) int {
	switch op {
	case OpPlus, OpMinus:
		return 1
	case OpMultiply, OpDivide:
		return 2
	default:
		return 0
	}
}

func (e *Executor) applyBinaryOp(a, b float64, op Operator) (float64, bool) {
	switch op {
	case OpPlus:
		return a + b, true
	case OpMinus:
		return a - b, true
	case OpMultiply:
		return a * b, true
	case OpDivide:
		return a / b, true
	default:
		return 0, false
	}
}

func (e *Executor) evaluate(tokens []Token) float64 {
	// stack to store integer values.
	var values utils.Stack[float64]

	// stack to store operators.
	var ops utils.Stack[Operator]

	for i := 0; i < len(tokens); i++ {
		if tokens[i].op == OpOpenParent { // Current token is an opening brace, push it to 'ops'
			ops.Push(tokens[i].op)
		} else if tokens[i].kind == KindNumber { // Current token is a number, push it to stack for numbers
			values.Push(tokens[i].number)
		} else if tokens[i].op == OpClosedParent { // Closing brace encountered, solve entire brace.
			for !ops.Empty() && ops.Top() != OpOpenParent {
				val2 := values.Pop()
				val1 := values.Pop()

				op := ops.Pop()

				res, ok := e.applyBinaryOp(val1, val2, op)
				if !ok {
					panic(ok)
				}
				values.Push(res)
			}

			if !ops.Empty() { // pop opening brace.
				ops.Pop()
			}
		} else {
			// While top of 'ops' has same or greater precedence to current token, which is an operator. Apply operator on top
			//of 'ops' to top two elements in values stack.
			for !ops.Empty() && e.precedence(ops.Top()) >= e.precedence(tokens[i].op) {
				val2 := values.Pop()
				val1 := values.Pop()

				op := ops.Pop()

				res, ok := e.applyBinaryOp(val1, val2, op)
				if !ok {
					panic(ok)
				}
				values.Push(res)
			}

			ops.Push(tokens[i].op) // Push current token to 'ops'.
		}
	}

	// Entire expression has been parsed at this point, apply remaining ops to remaining values.
	for !ops.Empty() {
		val2 := values.Pop()
		val1 := values.Pop()

		op := ops.Pop()

		res, ok := e.applyBinaryOp(val1, val2, op)
		if !ok {
			panic(ok)
		}
		values.Push(res)
	}

	return values.Top() // Top of 'values' contains result, return it.
}
