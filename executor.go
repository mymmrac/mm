package main

import (
	"fmt"
	"regexp"
	"strings"
)

type exprError struct {
	text string
	pos  int
}

func newExprErr(text string, pos int) *exprError {
	return &exprError{
		text: text,
		pos:  pos,
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
	tokens := e.lexer.tokenize(expr)
	return strings.Join(mapSlice(tokens, toString[token]), ", "), nil
}

type tokenKind string

type token struct {
	kind     tokenKind
	value    string
	startPos int
	endPos   int
}

func (t token) String() string {
	return fmt.Sprintf("{%s}:%d-%d `%s`", t.kind, t.startPos, t.endPos, t.value)
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
)

func (l *lexer) tokenize(text string) []token {
	var tokens []token

	var (
		pos     = 0
		trimmed int
		loc     []int
		tKind   tokenKind
		tValue  string
	)

	for text != "" {
		text, trimmed = trimWhitespacesAndCount(text)
		pos += trimmed

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
			panic("TODO: Handle invalid syntax")
		}

		tValue = text[loc[0]:loc[1]]

		tokens = append(tokens, token{
			kind:     tKind,
			value:    tValue,
			startPos: loc[0] + pos,
			endPos:   loc[1] + pos,
		})

		pos += len(tValue)
		text = text[loc[1]:]
	}

	return tokens
}
