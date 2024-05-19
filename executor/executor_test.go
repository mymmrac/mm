package executor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mymmrac/mm/debugger"
	"github.com/mymmrac/mm/executor"
)

func TestExecute(t *testing.T) {
	testcases := map[string]struct {
		expr   string
		result string
		err    bool
	}{
		"empty":             {expr: "", result: "", err: false},
		"number_int":        {expr: "123", result: "123", err: false},
		"number_float":      {expr: "1.23", result: "1.23", err: false},
		"add_int":           {expr: "1+2", result: "3", err: false},
		"add_float":         {expr: "1.1+2.2", result: "3.3", err: false},
		"subtract_int":      {expr: "2-1", result: "1", err: false},
		"subtract_float":    {expr: "1.1-2.2", result: "-1.1", err: false},
		"multiply_int":      {expr: "2*3", result: "6", err: false},
		"multiply_float":    {expr: "1.1*2.2", result: "2.42", err: false},
		"divide_int":        {expr: "6/3", result: "2", err: false},
		"divide_float":      {expr: "1.1/2.2", result: "0.5", err: false},
		"mod":               {expr: "11%3", result: "2", err: false},
		"abs":               {expr: "abs(-1)", result: "1", err: false},
		"sqr_root":          {expr: "sqrt(4)", result: "2", err: false},
		"sin":               {expr: "sin(0)", result: "0", err: false},
		"sin_negative":      {expr: "sin(-1)", result: "-0.8414709848078965", err: false},
		"divide_by_zero":    {expr: "1/0", result: "", err: true},
		"invalid_expr1":     {expr: "abc", result: "", err: true},
		"invalid_expr2":     {expr: "+-3", result: "", err: true},
		"unknown_func":      {expr: "abcdef(123)", result: "", err: true},
		"power":             {expr: "2^3", result: "8", err: false},
		"priority":          {expr: "(3+2)*2", result: "10", err: false},
		"grouping":          {expr: "(3+2)*2", result: "10", err: false},
		"group_parentheses": {expr: "(3+2)*2", result: "10", err: false},
		"group_braces":      {expr: "(3+(2))*2", result: "10", err: false},
		"sin_abs":           {expr: "abs(sin(-Pi/2))", result: "1", err: false},
		"sqrt_sqrt":         {expr: "sqrt(sqrt(16))", result: "2", err: false},
		"sqrt_zero":         {expr: "sqrt(0)", result: "0", err: false},
		"sqrt_negative":     {expr: "sqrt(-9)", result: "", err: true},
		"atan_two":          {expr: "atan(2)", result: "1.1071487177940905", err: false},
		"atan_half":         {expr: "atan(0.5)", result: "0.4636476090008061", err: false},
		"sin_half":          {expr: "1/sin(0.5)", result: "2.0858296429334882", err: false},
		"tan_half":          {expr: "1/tan(0.5)", result: "1.830487721712452", err: false},
		"atan_two_pi":       {expr: "atan(2*Pi)", result: "1.4129651365067377", err: false},
	}
	e := executor.NewExecutor(&debugger.Debugger{})
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			result, err := e.Execute(tc.expr, 16)
			if tc.err {
				assert.Error(t, err)
				assert.Equal(t, "", result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.result, result)
			}
		})
	}
}
