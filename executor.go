package main

import (
	"fmt"
	"regexp"
	"strings"
)

type location struct {
	start int
	end   int
}

func (l location) size() int {
	return l.end - l.start
}

type exprError struct {
	text string
	loc  location
}

func newExprErr(text string, loc location) *exprError {
	return &exprError{
		text: text,
		loc:  loc,
	}
}

type executor struct {
	lexer *lexer
}

func newExecutor() *executor {
	return &executor{
		lexer: newLexer(),
	}
}

func (e *executor) execute(expr string) (string, *exprError) {
	tokens, err := e.lexer.tokenize(expr)
	if err != nil {
		return "", err
	}

	return strings.Join(mapSlice(tokens, toString[token]), ", "), nil
}

type tokenKind string

type token struct {
	kind  tokenKind
	value string
	loc   location
}

func (t token) String() string {
	return fmt.Sprintf("{%s}:%d-%d `%s`", t.kind, t.loc.start, t.loc.end, t.value)
}

type lexer struct{}

func newLexer() *lexer {
	return &lexer{}
}

const (
	identifier tokenKind = "identifier" // `abc`, `a12`, `a_b_1`
	number     tokenKind = "number"     // `123`, `1.12`, `-12`, `1_2_3`
	operator   tokenKind = "operator"   // `+`, `-`, `^`, `(`
)

var (
	identPattern    = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`)
	numberPattern   = regexp.MustCompile(`^-?[0-9_]+(:?\.[0-9_]+)?`)
	operatorPattern = regexp.MustCompile(`^[-+*/^()]`)

	unknownPattern = regexp.MustCompile(`^[^ \t]+`)
)

func (l *lexer) tokenize(text string) ([]token, *exprError) {
	var tokens []token

	var (
		offset  = 0
		trimmed int
		loc     []int
		tKind   tokenKind
		tValue  string
	)

	for text != "" {
		text, trimmed = trimWhitespacesAndCount(text)
		offset += trimmed

		loc = identPattern.FindStringIndex(text)
		tKind = identifier

		if len(loc) == 0 {
			loc = numberPattern.FindStringIndex(text)
			tKind = number
		}

		if len(loc) == 0 {
			loc = operatorPattern.FindStringIndex(text)
			tKind = operator
		}

		if len(loc) == 0 {
			loc = unknownPattern.FindStringIndex(text)

			if len(loc) == 0 {
				return nil, newExprErr("invalid expression", location{start: 0, end: len(text) + offset})
			}

			return nil, newExprErr(
				fmt.Sprintf("unknown token: `%s`", text[loc[0]:loc[1]]),
				location{
					start: loc[0] + offset,
					end:   loc[1] + offset,
				},
			)
		}

		tValue = text[loc[0]:loc[1]]

		tokens = append(tokens, token{
			kind:  tKind,
			value: tValue,
			loc: location{
				start: loc[0] + offset,
				end:   loc[1] + offset,
			},
		})

		offset += len(tValue)
		text = text[loc[1]:]
	}

	return tokens, nil
}
