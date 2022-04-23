package main

import (
	"fmt"
	"strconv"

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

func (e *Executor) Execute(expr string) (string, *ExprError) {
	e.debugger.Clean()

	tokens, err := e.lexer.Tokenize(expr)
	if err != nil {
		return "", err
	}
	if len(tokens) == 0 {
		return "", nil
	}

	variables := make(map[Token]decimal.Decimal)

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

func (e *Executor) typeCheck(tokens []Token, variables map[Token]decimal.Decimal) *ExprError {
	var openParents utils.Stack[int]

	// Identify tokens
	for i, token := range tokens {
		switch token.kind {
		case KindIdentifier:
			value, ok := constants[token.text]
			if !ok {
				return NewExprErr("unknown identifier `"+token.text+"`", token.loc)
			}
			variables[token] = value
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

			// Validate parents
			if op == OpOpenParent {
				openParents.Push(i)
			} else if op == OpCloseParent {
				if openParents.Empty() {
					return NewExprErr("unexpected closing parent", token.loc)
				}
				openParents.Pop()
			}
		default:
			utils.Assert(false, "unreachable")
		}
	}

	// Validate parents
	if !openParents.Empty() {
		return NewExprErr("unclosed opened parent", tokens[openParents.Top()].loc)
	}

	// Type check operators arguments
	values := 0
	ops := utils.Stack[int]{}

	validate := func() bool {
		i := ops.Pop()
		op := tokens[i].op

		e.debugger.Debug(values, " ", opsToText[op], " ", tokens[i].loc)

		switch opsTypes[op] {
		case TypeUnary:
			if values < 1 {
				return false
			}
		case TypeBinary:
			if values < 2 {
				// Convert minus into unary minus
				if op == OpMinus && values == 1 {
					tokens[i].op = OpUnaryMinus
					return true
				}

				return false
			}
			values -= 1
		default:
			utils.Assert(false, "unreachable")
		}

		return true
	}

	parentValues := utils.Stack[int]{}

	for i, token := range tokens {
		if token.op == OpOpenParent {
			ops.Push(i)
			parentValues.Push(values)
		} else if token.kind == KindNumber || token.kind == KindIdentifier {
			values++
		} else if token.op == OpCloseParent {
			e.debugger.Debug(parentValues)

			beforeParents := parentValues.Pop()
			values -= beforeParents

			for !ops.Empty() && tokens[ops.Top()].op != OpOpenParent {
				opToken := tokens[ops.Top()]
				if ok := validate(); !ok {
					return NewExprErr("not enough args for "+opsToText[opToken.op]+" operation", opToken.loc)
				}
			}

			values = beforeParents + 1

			if !ops.Empty() {
				ops.Pop()
			}
		} else {
			for !ops.Empty() && compareOpPrecedence(tokens[ops.Top()], token) {
				opToken := tokens[ops.Top()]
				if ok := validate(); !ok {
					return NewExprErr("not enough args for "+opsToText[opToken.op]+" operation", opToken.loc)
				}
			}

			ops.Push(i)
		}
	}

	for !ops.Empty() {
		opToken := tokens[ops.Top()]
		if ok := validate(); !ok {
			return NewExprErr("not enough args for "+opsToText[opToken.op]+" operation", opToken.loc)
		}
	}

	return nil
}

func (e *Executor) evaluate(tokens []Token, variables map[Token]decimal.Decimal) (Token, *ExprError) {
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
