package main

import (
	"strconv"

	"github.com/mymmrac/mm/utils"
)

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

	result := e.evaluate(tokens)
	if result.kind != KindNumber {
		return "", NewExprErr("returned invalid result type: "+result.String(), result.loc)
	}

	return strconv.FormatFloat(result.number, 'f', -1, 64), nil
}

func (e *Executor) typeCheck(tokens []Token) *ExprError {
	for i, token := range tokens {
		switch token.kind {
		case KindIdentifier:
			return NewExprErr("identifiers not supported yet", token.loc)
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

func (e *Executor) evaluate(tokens []Token) Token {
	var values utils.Stack[Token]
	var ops utils.Stack[Operator]

	for i := 0; i < len(tokens); i++ {
		if tokens[i].op == OpOpenParent {
			ops.Push(tokens[i].op)
		} else if tokens[i].kind == KindNumber {
			values.Push(tokens[i])
		} else if tokens[i].op == OpClosedParent {
			for !ops.Empty() && ops.Top() != OpOpenParent {
				val2 := values.Pop()
				val1 := values.Pop()

				op := ops.Pop()

				res, ok := applyBinaryOp(val1, val2, op)
				if !ok {
					panic(ok)
				}
				values.Push(res)
			}

			if !ops.Empty() {
				ops.Pop()
			}
		} else {
			for !ops.Empty() && opPrecedence(ops.Top()) >= opPrecedence(tokens[i].op) {
				val2 := values.Pop()
				val1 := values.Pop()

				op := ops.Pop()

				res, ok := applyBinaryOp(val1, val2, op)
				if !ok {
					panic(ok)
				}
				values.Push(res)
			}

			ops.Push(tokens[i].op)
		}
	}

	for !ops.Empty() {
		val2 := values.Pop()
		val1 := values.Pop()

		op := ops.Pop()

		res, ok := applyBinaryOp(val1, val2, op)
		if !ok {
			panic(ok)
		}
		values.Push(res)
	}

	return values.Top()
}
