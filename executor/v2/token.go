package executor

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Token struct {
	text string
	loc  Location
	kind TokenKind

	number     *decimal.Decimal
	operator   *Operator
	identifier *Identifier
}

func (t Token) String() string {
	s := fmt.Sprintf("{%s}:[%d-%d] `%s`", t.kind, t.loc.Start, t.loc.End, t.text)
	if t.number != nil {
		s += fmt.Sprintf(" %s", t.number.String())
	}
	if t.operator != nil {
		s += fmt.Sprintf(" %s", t.operator.name)
	}
	if t.identifier != nil {
		s += fmt.Sprintf(" %s", t.identifier.name)
		if !t.identifier.variable {
			s += fmt.Sprintf("/%d", t.identifier.arity)
		}
	}
	return s
}

type TokenKind string

const (
	KindNumber     TokenKind = "number"     // `123`, `1.12`, `12`, `1_2_3`
	KindOperator   TokenKind = "operator"   // `+`, `-`, `//`, `(`
	KindIdentifier TokenKind = "identifier" // `abc`, `a12`, `a_b_1`
)

type Location struct {
	Start int
	End   int
}

func (l Location) Size() int {
	return l.End - l.Start
}
