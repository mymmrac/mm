package executor

import (
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/mymmrac/mm/utils"
)

func (e *Executor) identifyTokens(tokens []Token, variables Vars) *ExprError {
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
			op, ok := textToOp[token.text]
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

	return nil
}

func (e *Executor) updateTokens(tokens []Token) {
	values := 0
	var parentValues, ops utils.Stack[int]

	// Update tokens types
	updateTypes := func() {
		i := ops.Pop()
		op := tokens[i].op

		switch opTypes[op] {
		case TypeUnary:
			// Do nothing
		case TypeBinary:
			// Convert minus into unary minus
			if op == OpMinus && values == 1 {
				tokens[i].op = OpUnaryMinus
				break
			}

			values--
		default:
			utils.Assert(false, "unreachable")
		}
	}

	for i, token := range tokens {
		if token.op == OpOpenParent {
			ops.Push(i)
			parentValues.Push(values)
		} else if token.kind == KindNumber || token.kind == KindIdentifier {
			values++
		} else if token.op == OpCloseParent {
			beforeParents := parentValues.Pop()
			values -= beforeParents

			for !ops.Empty() && tokens[ops.Top()].op != OpOpenParent {
				updateTypes()
			}

			values = beforeParents + 1

			if !ops.Empty() {
				ops.Pop()
			}
		} else {
			for !ops.Empty() && compareOpPrecedence(tokens[ops.Top()], token) {
				updateTypes()
			}

			ops.Push(i)
		}
	}

	for !ops.Empty() {
		updateTypes()
	}
}

func (e *Executor) validateTokens(tokens []Token) *ExprError {
	values := 0
	var parentValues, ops utils.Stack[int]

	// Type check operator's arguments
	validate := func() bool {
		i := ops.Pop()
		op := tokens[i].op

		switch opTypes[op] {
		case TypeUnary:
			if values < 1 {
				return false
			}
		case TypeBinary:
			if values < 2 {
				return false
			}
			values--
		default:
			utils.Assert(false, "unreachable")
		}

		return true
	}

	isValue := func(kind TokenKind) bool {
		return kind == KindNumber || kind == KindIdentifier
	}

	for i, token := range tokens {
		if token.op == OpOpenParent {
			// Check order of values
			if i > 0 && isValue(tokens[i-1].kind) {
				return NewExprErr("no operator found for `"+token.text+"`", token.loc)
			}

			ops.Push(i)
			parentValues.Push(values)
		} else if isValue(token.kind) {
			// Check order of values
			if i > 0 && isValue(tokens[i-1].kind) {
				return NewExprErr("no operator found for `"+token.text+"`", token.loc)
			}

			values++
		} else if token.op == OpCloseParent {
			beforeParents := parentValues.Pop()
			values -= beforeParents

			for !ops.Empty() && tokens[ops.Top()].op != OpOpenParent {
				opToken := tokens[ops.Top()]
				if ok := validate(); !ok {
					return NewExprErr("not enough args for "+opToText[opToken.op]+" operation", opToken.loc)
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
					return NewExprErr("not enough args for "+opToText[opToken.op]+" operation", opToken.loc)
				}
			}

			ops.Push(i)
		}
	}

	for !ops.Empty() {
		opToken := tokens[ops.Top()]
		if ok := validate(); !ok {
			return NewExprErr("not enough args for "+opToText[opToken.op]+" operation", opToken.loc)
		}
	}

	return nil
}
