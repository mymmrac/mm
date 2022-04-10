package main

import (
	"strconv"

	"github.com/mymmrac/mm/utils"
	"github.com/shopspring/decimal"
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
	if len(tokens) == 0 {
		return "", nil
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

	return result.number.String(), nil
}

func (e *Executor) typeCheck(tokens []Token) *ExprError {
	var openParents utils.Stack[int]

	for i, token := range tokens {
		switch token.kind {
		case KindIdentifier:
			return NewExprErr("identifiers not supported yet", token.loc)
		case KindNumber:
			n, err := strconv.ParseFloat(token.text, 64)
			if err != nil {
				return NewExprErr("parsing number: "+err.Error(), token.loc)
			}
			tokens[i].number = decimal.NewFromFloat(n)
		case KindOperator:
			op, ok := textToOps[token.text]
			if !ok {
				return NewExprErr("unknown operator: "+token.text, token.loc)
			}
			tokens[i].op = op

			if op == OpOpenParent {
				openParents.Push(i)
			} else if op == OpClosedParent {
				if openParents.Empty() {
					return NewExprErr("unexpected closing parent", token.loc)
				}
				openParents.Pop()
			}
		default:
			utils.Assert(false, "unreachable")
		}
	}

	if !openParents.Empty() {
		return NewExprErr("unclosed opened parent", tokens[openParents.Top()].loc)
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
		opToken := ops.Pop()

		switch opsTypes[opToken.op] {
		case TypeUnary:
			if values.Size() < 1 {
				return NewExprErr("not enough args for `"+opsToText[opToken.op]+"` operation", opToken.loc)
			}

			v := values.Pop()

			res, ok := applyUnaryOp(v, opToken)
			if !ok {
				return NewExprErr("can't apply `"+opsToText[opToken.op]+"` operation", res.loc)
			}

			values.Push(res)
		case TypeBinary:
			if values.Size() < 2 {
				return NewExprErr("not enough args for `"+opsToText[opToken.op]+"` operation", opToken.loc)
			}

			v2 := values.Pop()
			v1 := values.Pop()

			res, ok := applyBinaryOp(v1, v2, opToken)
			if !ok {
				return NewExprErr("can't apply `"+opsToText[opToken.op]+"` operation", res.loc)
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
			for !ops.Empty() && compareOpPrecedence(ops.Top(), token) {
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

func compareOpPrecedence(op1, op2 Token) bool {
	if opPrecedence(op1.op) != opPrecedence(op2.op) {
		return opPrecedence(op1.op) > opPrecedence(op2.op)
	}
	return op1.loc.start > op2.loc.start
}
