package main

import (
	"fmt"

	"github.com/shopspring/decimal"
)

var one = decimal.New(1, 0)

func DecimalRoot(d1, d2 decimal.Decimal) decimal.Decimal {
	if d1.IsNegative() {
		panic(fmt.Sprintf("d1 in root is negative: %s", d1))
	}
	if !d2.IsInteger() {
		panic(fmt.Sprintf("d2 in root is not integer: %s", d2))
	}

	if d2.IsNegative() {
		d1 = one.Div(d1)
		d2 = d2.Neg()
	}

	n1 := d2.Sub(one)

	a := n1.Div(d2)
	b := d1.Div(d2)

	x := b

	for i := 0; i < 256; i++ {
		x = a.Mul(x).Add(b.Div(x.Pow(n1)))
	}

	return x.Round(16)
}
