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

	result, err := e.evaluate(tokens)
	if err != nil {
		return "", err
	}
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

func (e *Executor) evaluate(tokens []Token) (Token, *ExprError) {
	var values utils.Stack[Token]
	var ops utils.Stack[Operator]

	eval := func() *ExprError {
		op := ops.Pop()

		val2 := values.Pop()
		val1 := values.Pop()

		res, ok := applyBinaryOp(val1, val2, op)
		if !ok {
			return NewExprErr("can't apply `"+opsToText[op]+"` operation", res.loc)
		}

		values.Push(res)
		return nil
	}

	for _, token := range tokens {
		if token.op == OpOpenParent {
			ops.Push(token.op)
		} else if token.kind == KindNumber {
			values.Push(token)
		} else if token.op == OpClosedParent {
			for !ops.Empty() && ops.Top() != OpOpenParent {
				if err := eval(); err != nil {
					return Token{}, err
				}
			}

			if !ops.Empty() {
				ops.Pop()
			}
		} else {
			for !ops.Empty() && opPrecedence(ops.Top()) >= opPrecedence(token.op) {
				if err := eval(); err != nil {
					return Token{}, err
				}
			}

			ops.Push(token.op)
		}
	}

	for !ops.Empty() {
		if err := eval(); err != nil {
			return Token{}, err
		}
	}

	switch values.Size() {
	case 0:
		return Token{}, NewExprErr("no values left", Location{})
	case 1:
		return values.Top(), nil
	default:
		values.Pop()
		return Token{}, NewExprErr("not handled value left", values.Top().loc)
	}
}
