package main

import (
	"fmt"

	"github.com/shopspring/decimal"

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

func (e *ExprError) Error() string {
	if e.loc.Size() == 1 {
		return fmt.Sprintf("expression at [%d]: %s", e.loc.start+1, e.text)
	}
	return fmt.Sprintf("expression in rage [%d, %d]: %s", e.loc.start+1, e.loc.end, e.text)
}

type Executor struct {
	lexer    *Lexer
	debugger *Debugger
}

func NewExecutor(debugger *Debugger) *Executor {
	return &Executor{
		lexer:    NewLexer(),
		debugger: debugger,
	}
}

type Vars map[Token]decimal.Decimal

func (e *Executor) Execute(expr string) (string, *ExprError) {
	e.debugger.Clean()

	tokens, err := e.lexer.Tokenize(expr)
	if err != nil {
		return "", err
	}
	if len(tokens) == 0 {
		return "", nil
	}

	variables := make(Vars)

	if err = e.typeCheck(tokens, variables); err != nil {
		return "", err
	}

	e.debugger.Debug("Tokens ", tokens)

	result, err := e.evaluate(tokens, variables)
	if err != nil {
		return "", err
	}

	switch result.kind {
	case KindNumber:
		return result.number.String(), nil
	case KindIdentifier:
		number, ok := variables[result]
		if !ok {
			return "", NewExprErr("returned unknown identifier: "+result.String(), result.loc)
		}
		return number.String(), nil
	default:
		return "", NewExprErr("returned invalid result type: "+result.String(), result.loc)
	}
}

func (e *Executor) typeCheck(tokens []Token, variables Vars) *ExprError {
	e.debugger.Debug("Raw ", tokens)

	if err := e.identifyTokens(tokens, variables); err != nil {
		return err
	}

	e.debugger.Debug("Identified ", tokens)

	if err := e.updateTokens(tokens); err != nil {
		return err
	}

	e.debugger.Debug("Type checked ", tokens)

	if err := e.validateTokens(tokens); err != nil {
		return err
	}

	return nil
}

func (e *Executor) evaluate(tokens []Token, variables Vars) (Token, *ExprError) {
	var values, ops utils.Stack[Token]

	eval := func() *ExprError {
		opToken := ops.Pop()

		switch opsTypes[opToken.op] {
		case TypeUnary:
			if values.Size() < 1 {
				return NewExprErr("not enough args for `"+opsToText[opToken.op]+"` operation", opToken.loc)
			}

			v := values.Pop()

			res, ok := applyUnaryOp(v, opToken, variables)
			if !ok {
				return NewExprErr("can't apply "+opsToText[opToken.op]+" operation", res.loc)
			}

			values.Push(res)
		case TypeBinary:
			if values.Size() < 2 {
				return NewExprErr("not enough args for "+opsToText[opToken.op]+" operation", opToken.loc)
			}

			v2 := values.Pop()
			v1 := values.Pop()

			res, ok := applyBinaryOp(v1, v2, opToken, variables)
			if !ok {
				return NewExprErr("can't apply "+opsToText[opToken.op]+" operation", res.loc)
			}

			values.Push(res)
		default:
			return NewExprErr(fmt.Sprintf("unkown type of `%s` operation", opToken.text), opToken.loc)
		}

		return nil
	}

	for _, token := range tokens {
		if token.op == OpOpenParent {
			ops.Push(token)
		} else if token.kind == KindNumber || token.kind == KindIdentifier {
			values.Push(token)
		} else if token.op == OpCloseParent {
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
