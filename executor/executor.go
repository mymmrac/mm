package executor

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/mymmrac/mm/debugger"
	"github.com/mymmrac/mm/utils"
)

type Executor struct {
	debugger *debugger.Debugger
}

func NewExecutor(debugger *debugger.Debugger) *Executor {
	return &Executor{
		debugger: debugger,
	}
}

func (e *Executor) Execute(expression string, precision int32) (string, error) {
	e.debugger.Clean()

	tokens, err := e.tokenize(expression)
	if err != nil {
		return "", err
	}
	e.debugger.Debug("Tokens ", tokens)

	if len(tokens) == 0 {
		return "", nil
	}

	err = e.typeCheck(tokens)
	if err != nil {
		return "", err
	}
	e.debugger.Debug("Tokens (type checked) ", tokens)

	tokens, err = e.convertToPostfixNotation(tokens)
	if err != nil {
		return "", err
	}
	e.debugger.Debug("Tokens (postfix notation) ", tokens)

	result, err := e.evaluate(tokens)
	if err != nil {
		return "", err
	}

	return result.Round(precision).String(), nil
}

func (e *Executor) tokenize(expression string) ([]Token, error) {
	i := 0
	var tokens []Token
	for i < len(expression) {
		if utils.IsSpace(expression[i]) {
			i++
			continue
		}

		if utils.IsDigit(expression[i]) {
			j := i + 1

			hasDot := false
			hasExp := false
			hasExpSign := false
			expSignPos := -1

		numberLoop:
			for j < len(expression) {
				switch {
				case utils.IsDigit(expression[j]):
					j++
				case !hasDot && expression[j] == '.':
					hasDot = true
					j++
				case !hasExpSign && utils.IsInCharset(expression[j], "+-"):
					hasExpSign = true
					expSignPos = j
					j++
				case !hasExp && utils.IsInCharset(expression[j], "eE"):
					hasExp = true
					j++
				default:
					break numberLoop
				}
			}

			if hasExpSign && !hasExp {
				j = expSignPos
			}

			tokens = append(tokens, Token{
				text: expression[i:j],
				loc: Location{
					Start: i,
					End:   j,
				},
				kind: KindNumber,
			})
			i = j
			continue
		}

		opIndex := slices.IndexFunc(knownUniqueOperators, func(op string) bool {
			if len(op)+i > len(expression) {
				return false
			}
			return expression[i:i+len(op)] == op
		})
		if opIndex != -1 {
			op := knownUniqueOperators[opIndex]
			tokens = append(tokens, Token{
				text: op,
				kind: KindOperator,
				loc: Location{
					Start: i,
					End:   i + len(op),
				},
			})
			i += len(op)
			continue
		}

		identIndex := slices.IndexFunc(knownUniqueIdentifiers, func(ident string) bool {
			if len(ident)+i > len(expression) {
				return false
			}
			return expression[i:i+len(ident)] == ident
		})
		if identIndex != -1 {
			ident := knownUniqueIdentifiers[identIndex]
			tokens = append(tokens, Token{
				text: ident,
				kind: KindIdentifier,
				loc: Location{
					Start: i,
					End:   i + len(ident),
				},
			})
			i += len(ident)
			continue
		}

		return nil, NewExprError("invalid symbol", Location{Start: i, End: i + 1})
	}
	return tokens, nil
}

func (e *Executor) typeCheck(tokens []Token) error {
	lValues := 0
	lastLValue := -1
	openParents := 0
	lastOpenParentIndex := -1

	for i, token := range tokens {
		switch token.kind {
		case KindNumber:
			number, err := decimal.NewFromString(token.text)
			if err != nil {
				return fmt.Errorf("parse number: %w", err)
			}
			tokens[i].number = &number
			lValues++
			lastLValue = i
		case KindOperator:
			switch token.text {
			case opOpenParenthesis.text:
				openParents++
				lastOpenParentIndex = i
				tokens[i].operator = &opOpenParenthesis
			case opCloseParenthesis.text:
				if openParents == 0 {
					return NewExprError("unexpected closing parenthesis", token.loc)
				}

				if tokens[i-1].isOpenParenthesis() && (i == 1 || (tokens[i-2].kind != KindIdentifier ||
					tokens[i-2].identifier.arity != 0 || tokens[i-2].identifier.variable)) {
					return NewExprError("unexpected closing parenthesis", token.loc)
				}

				openParents--
				tokens[i].operator = &opCloseParenthesis
			case opComma.text:
				// TODO: Check tha only used inside functions
				tokens[i].operator = &opComma
			default:
				if i > 0 {
					pt := tokens[i-1]
					if pt.kind == KindOperator && !pt.isControlFlow() {
						return NewExprError("unexpected operator `"+token.text+"`", token.loc)
					}
				}

				opIndex := -1

				if i == 0 || (tokens[i-1].kind == KindOperator && tokens[i-1].text != opCloseParenthesis.text) {
					opIndex = slices.IndexFunc(knownOperators, func(op Operator) bool {
						return op.arity == 1 && op.text == token.text
					})
				} else {
					opIndex = slices.IndexFunc(knownOperators, func(op Operator) bool {
						return op.arity == 2 && op.text == token.text
					})
					lValues--
				}

				if opIndex < 0 {
					return NewExprError("unknown operator `"+token.text+"`", token.loc)
				}

				tokens[i].operator = &knownOperators[opIndex]
			}
		case KindIdentifier:
			identIndex := -1

			if i == len(tokens)-1 || !(tokens[i+1].isOpenParenthesis()) {
				identIndex = slices.IndexFunc(knownIdentifiers, func(ident Identifier) bool {
					return ident.variable && ident.text == token.text
				})
				if identIndex < 0 {
					return NewExprError("unknown identifier `"+token.text+"`", token.loc)
				}
				lValues++
				lastLValue = i
			} else {
				parenthesis := 0
				var args uint = 0
				capturingArg := false
				var unusedComma *Token
				for k, t := range tokens[i+1:] {
					// Check first open parenthesis
					if k == 0 {
						if !t.isOpenParenthesis() {
							return NewExprError(
								"expected `"+opOpenParenthesis.text+"`, but got `"+t.text+"`",
								t.loc,
							)
						}
						parenthesis++
						continue
					}

					// Add open parenthesis
					if t.isOpenParenthesis() {
						if !capturingArg {
							args++
							capturingArg = true
							unusedComma = nil
						}
						parenthesis++
						continue
					}
					// Remove open parenthesis
					if t.isCloseParenthesis() {
						parenthesis--
						if parenthesis == 0 {
							break
						}
						continue
					}

					// Skip everything inside parenthesis
					if parenthesis > 1 {
						continue
					}

					// Check comma
					if t.isComma() {
						unusedComma = &t
						capturingArg = false
						continue
					}

					// Add argument
					if !capturingArg {
						args++
						capturingArg = true
						unusedComma = nil
					}
				}

				if unusedComma != nil {
					return NewExprError("unexpected "+opComma.name, unusedComma.loc)
				}

				identIndex = slices.IndexFunc(knownIdentifiers, func(ident Identifier) bool {
					return !ident.variable && ident.arity == args && ident.text == token.text
				})
				if identIndex < 0 {
					return NewExprError(
						"unknown function `"+token.text+"/"+strconv.FormatUint(uint64(args), 10)+"`",
						token.loc,
					)
				}

				lValues -= int(args)
				lValues++
				lastLValue = i
			}

			tokens[i].identifier = &knownIdentifiers[identIndex]
		default:
			return NewExprError(fmt.Sprintf("unknown token kind: %q", token.kind), token.loc)
		}
	}

	if openParents != 0 {
		return NewExprError("unexpected opening parenthesis", tokens[lastOpenParentIndex].loc)
	}
	switch {
	case lValues < 0:
		return NewExprError("too many values consumed in expression", Location{
			Start: tokens[0].loc.Start,
			End:   tokens[len(tokens)-1].loc.End,
		})
	case lValues == 0:
		return NewExprError("no values returned in expression", Location{
			Start: tokens[0].loc.Start,
			End:   tokens[len(tokens)-1].loc.End,
		})
	case lValues == 1:
		return nil
	default:
		// FIXME: last L value for functions with arguments is not correct
		return NewExprError("too many values returned in expression", tokens[lastLValue].loc)
	}
}

func (e *Executor) convertToPostfixNotation(tokens []Token) ([]Token, error) {
	stack := utils.NewStack[Token]()
	output := utils.NewStack[Token]()

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		switch token.kind {
		case KindNumber:
			output.Push(token)
		case KindOperator:
			switch token.text {
			case opOpenParenthesis.text:
				stack.Push(token)
			case opCloseParenthesis.text:
				for !stack.Empty() && !stack.Top().isOpenParenthesis() {
					output.Push(stack.Pop())
				}
				_ = stack.Pop()
			case opComma.text:
				return nil, NewExprError("unexpected `"+opComma.text+"`", token.loc)
			default:
				for !stack.Empty() && !stack.Top().isOpenParenthesis() &&
					token.operator.precedence <= stack.Top().operator.precedence {
					output.Push(stack.Pop())
				}
				stack.Push(token)
			}
		case KindIdentifier:
			if token.identifier.variable {
				output.Push(token)
			} else {
				i += 2 // Skip identifier and open parenthesis
				for j := uint(0); j < token.identifier.arity; j++ {
					si := i

					t := tokens[i]
					if j != token.identifier.arity-1 {
						for !t.isComma() {
							i++
							t = tokens[i]
						}
					} else {
						p := 0
						for {
							switch {
							case t.isOpenParenthesis():
								p++
							case t.isCloseParenthesis():
								p--
							}
							if p == -1 {
								break
							}

							i++
							t = tokens[i]
						}
					}

					arg, err := e.convertToPostfixNotation(tokens[si:i])
					if err != nil {
						return nil, err
					}
					if j != token.identifier.arity-1 {
						i++ // Skip comma
					}

					output.Push(arg...)
				}
				// Skip close parenthesis (current token) done in the loop above
				output.Push(token)
			}
		default:
			return nil, NewExprError(fmt.Sprintf("unknown token kind: %q", token.kind), token.loc)
		}
	}

	for !stack.Empty() {
		output.Push(stack.Pop())
	}

	return output.Slice(), nil
}

func (e *Executor) evaluate(tokens []Token) (decimal.Decimal, error) {
	stack := utils.NewStack[decimal.Decimal]()

	for _, token := range tokens {
		switch token.kind {
		case KindNumber:
			stack.Push(*token.number)
		case KindOperator:
			if err := token.operator.apply(stack); err != nil {
				return decimal.Zero, NewExprError(
					fmt.Sprintf("apply operator `"+token.text+"`: %s", err),
					token.loc,
				)
			}
		case KindIdentifier:
			if err := token.identifier.apply(stack); err != nil {
				identType := "function"
				if token.identifier.variable {
					identType = "variable"
				}
				return decimal.Zero, NewExprError(
					fmt.Sprintf("apply %s `"+token.text+"`: %s", identType, err),
					token.loc,
				)
			}
		default:
			return decimal.Zero, NewExprError(fmt.Sprintf("unknown token kind: %q", token.kind), token.loc)
		}
	}

	switch stack.Size() {
	case 0:
		return decimal.Zero, fmt.Errorf("no return values")
	case 1:
		return stack.Pop(), nil
	default:
		return decimal.Zero, fmt.Errorf("too many (%d) values returned", stack.Size())
	}
}
