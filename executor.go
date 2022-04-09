package main

import "strings"

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

type executor struct{}

func newExecutor() *executor {
	return &executor{}
}

func (e *executor) execute(expr string) (string, *exprError) {
	i := strings.Index(expr, "e")
	if i >= 0 {
		return "", newExprErr("found `e`", i)
	}

	return expr, nil
}
