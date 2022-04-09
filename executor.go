package main

import (
	"fmt"
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

type lexer struct{}

func newLexer() *lexer {
	return &lexer{}
}

func (l *lexer) tokenize(text string) []token {
	return nil
}

type tokenKind string

const (
	identifier tokenKind = "identifier" // `abc`, `a12`, `a_b_1`
	number     tokenKind = "number"     // `123`, `1.12`, `-12`, `1_2_3`
	operator   tokenKind = "operator"   // `+`, `-`, `^`
)

type token struct {
	kind  tokenKind
	value string
	pos   int
}

func (t token) String() string {
	return fmt.Sprintf("{%s}:%d `%s`", t.kind, t.pos, t.value)
}
