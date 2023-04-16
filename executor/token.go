package executor

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Token struct {
	kind   TokenKind
	text   string
	loc    Location
	number decimal.Decimal
	op     Operator
}

func (t Token) String() string {
	return fmt.Sprintf("{%s}:[%d-%d] `%s` %s %q", t.kind, t.loc.Start, t.loc.End, t.text, t.number, opToText[t.op])
}

type TokenKind string

const (
	KindIdentifier TokenKind = "identifier" // `abc`, `a12`, `a_b_1`
	KindNumber     TokenKind = "number"     // `123`, `1.12`, `12`, `1_2_3`
	KindOperator   TokenKind = "operator"   // `+`, `-`, `^`, `(`
)

type Location struct {
	Start int
	End   int
}

func (l Location) Size() int {
	return l.End - l.Start
}

type Vars map[Token]decimal.Decimal
