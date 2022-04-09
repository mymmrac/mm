package main

import (
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
