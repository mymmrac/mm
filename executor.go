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

	for i, token := range tokens {
		if token.kind != KindOperator || token.op != OpMinus {
			continue
		}

		// - ...
		if i == 0 {
			tokens[i].op = OpUnaryMinus
			continue
		}

		// ... ( - ...
		if i > 0 && tokens[i-1].kind == KindOperator && tokens[i-1].op == OpOpenParent {
			tokens[i].op = OpUnaryMinus
		}
	}

	return nil
}

func (e *Executor) evaluate(tokens []Token) (Token, *ExprError) {
	var values, ops utils.Stack[Token]

	eval := func() *ExprError {
		op := ops.Pop()

		switch opsTypes[op.op] {
		case TypeUnary:
			v := values.Pop()

			res, ok := applyUnaryOp(v, op)
			if !ok {
				return NewExprErr("can't apply `"+opsToText[op.op]+"` operation", res.loc)
			}

			values.Push(res)
		case TypeBinary:
			v2 := values.Pop()
			v1 := values.Pop()

			res, ok := applyBinaryOp(v1, v2, op)
			if !ok {
				return NewExprErr("can't apply `"+opsToText[op.op]+"` operation", res.loc)
			}

			values.Push(res)
		}

		return nil
	}

	for _, token := range tokens {
		if token.op == OpOpenParent {
			ops.Push(token)
		} else if token.kind == KindNumber {
			values.Push(token)
		} else if token.op == OpClosedParent {
			for !ops.Empty() && ops.Top().op != OpOpenParent {
				if err := eval(); err != nil {
					return Token{}, err
				}
			}

			if !ops.Empty() {
				ops.Pop()
			}
		} else {
			for !ops.Empty() && opPrecedence(ops.Top().op) >= opPrecedence(token.op) {
				if err := eval(); err != nil {
					return Token{}, err
				}
			}

			ops.Push(token)
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
