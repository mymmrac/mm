package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func l(t *testing.T, s, e int) Location {
	t.Helper()
	return Location{
		start: s,
		end:   e,
	}
}

func nl(t *testing.T) Location {
	t.Helper()
	return Location{
		start: -1,
		end:   -1,
	}
}

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name   string
		expr   string
		result string
		loc    Location
	}{
		{name: "empty", expr: "", result: "", loc: nl(t)},
		{name: "empty_with_whitespaces", expr: " \t \t  ", result: "", loc: nl(t)},

		{name: "number", expr: "123", result: "123", loc: nl(t)},
		{name: "negative_number", expr: "-123", result: "-123", loc: nl(t)},
		{name: "float_number", expr: "12.34", result: "12.34", loc: nl(t)},
		{name: "negative_float_number", expr: "-12.34", result: "-12.34", loc: nl(t)},
		{name: "complex_number", expr: "-123_123_31.1_1_23_123", result: "-12312331.1123123", loc: nl(t)},

		{name: "number_with_parents", expr: " ( 1 )", result: "1", loc: nl(t)},
		{name: "number_with_multiple_parents", expr: " ( ( ( 1 )) )", result: "1", loc: nl(t)},

		{name: "err_number", expr: "1__123", result: "", loc: l(t, 0, 6)},
		{name: "err_number_with_whitespaces", expr: "   1__123  ", result: "", loc: l(t, 3, 9)},
		{name: "err_number_with_parents_unclosed", expr: " ( 1 ", result: "", loc: l(t, 1, 2)},
		{name: "err_number_with_multiple_parents_unclosed", expr: " ( ( (1)) ", result: "", loc: l(t, 1, 2)},
		{name: "err_number_with_multiple_parents_unopened", expr: "   ( (1 )) )", result: "", loc: l(t, 11, 12)},

		{name: "add", expr: "123.321 + 321.123", result: "444.444", loc: nl(t)},
		{name: "add_multiple", expr: "1 + 10 + 100 + 1000", result: "1111", loc: nl(t)},

		{name: "sub", expr: "123.321 - 321.123", result: "-197.802", loc: nl(t)},
		{name: "sub_compact_ expr_1", expr: "1-2", result: "-1", loc: nl(t)},
		{name: "sub_compact_ expr_2", expr: "1-(2)", result: "-1", loc: nl(t)},
		{name: "sub_compact_ expr_3", expr: "(1)-2", result: "-1", loc: nl(t)},

		{name: "unary_minus", expr: "- 321.123", result: "-321.123", loc: nl(t)},

		{name: "bug_unary_minus_1", expr: "1 + ( - 1 )", result: "0", loc: nl(t)},
		{name: "bug_unary_minus_2", expr: "1 + ( - 2 )", result: "-1", loc: nl(t)},

		{name: "expression_1", expr: "9 @ (3+1) + 17 / (6 - 12)", result: "-1.101282525764456", loc: nl(t)},

		//{name: "", expr: "", result: "", loc: nl(t)},
	}

	e := NewExecutor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.Execute(tt.expr)
			assert.Equal(t, tt.result, result)
			if tt.loc.start == -1 && tt.loc.end == -1 {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.True(t, len(err.text) > 0)
				assert.Equal(t, tt.loc, err.loc)
			}
		})
	}
}
