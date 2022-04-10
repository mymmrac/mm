package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mymmrac/mm/utils"
)

func init() {
	utils.Assert(int(OpsLast)-1 == len(textToOps), "ops count does not match", OpsLast-1, textToOps)
	utils.Assert(len(opsTypes) == len(textToOps), "ops count does not match ops type count", opsTypes, textToOps)

	opsToText = make(map[Operator]string, len(textToOps))
	for text, op := range textToOps {
		opsToText[op] = text
	}

	_, foundNoOp := opsToText[OpNoOp]
	utils.Assert(foundNoOp == false, "no-op found", opsToText)

	opsText := utils.Keys(textToOps)

	var singleCharOps, multiCharOps []string
	utils.ForeachSlice(opsText, func(op string) {
		if len(op) == 1 {
			singleCharOps = append(singleCharOps, op)
		} else {
			multiCharOps = append(multiCharOps, op)
		}
	})

	escapeAll := func(op string) string {
		escaped := ""
		for _, r := range op {
			escaped += fmt.Sprintf("\\%c", r)
		}
		return escaped
	}

	var opPattern string
	if len(multiCharOps) > 0 {
		opPattern = fmt.Sprintf(`^(:?%s|[%s])`,
			strings.Join(utils.MapSlice(multiCharOps, escapeAll), "|"),
			strings.Join(utils.MapSlice(singleCharOps, escapeAll), ""))
	} else {
		opPattern = fmt.Sprintf(`^[%s]`,
			strings.Join(utils.MapSlice(singleCharOps, escapeAll), ""))
	}
	operatorPattern = regexp.MustCompile(opPattern)
}

type Location struct {
	start int
	end   int
}

func (l Location) Size() int {
	return l.end - l.start
}

type Token struct {
	kind   TokenKind
	text   string
	loc    Location
	number float64
	op     Operator
}

func (t Token) String() string {
	return fmt.Sprintf("{%s}:%d-%d `%s`", t.kind, t.loc.start, t.loc.end, t.text)
}

type Lexer struct{}

func NewLexer() *Lexer {
	return &Lexer{}
}

var (
	identPattern    = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`)
	numberPattern   = regexp.MustCompile(`^-?[0-9_]+(:?\.[0-9_]+)?`)
	operatorPattern *regexp.Regexp // `^[-+*/^()]`

	unknownPattern = regexp.MustCompile(`^[^ \t]+`)
)

func (l *Lexer) Tokenize(text string) ([]Token, *ExprError) {
	var tokens []Token
	var (
		offset  = 0
		trimmed int
		loc     []int
		tKind   TokenKind
		tValue  string
	)

	text, trimmed = utils.TrimWhitespacesAndCount(text)
	offset += trimmed

	for text != "" {
		loc = identPattern.FindStringIndex(text)
		tKind = KindIdentifier

		if len(loc) == 0 {
			loc = numberPattern.FindStringIndex(text)
			tKind = KindNumber
		}

		if len(loc) == 0 {
			loc = operatorPattern.FindStringIndex(text)
			tKind = KindOperator
		}

		if len(loc) == 0 {
			loc = unknownPattern.FindStringIndex(text)

			if len(loc) == 0 {
				return nil, NewExprErr("invalid expression", Location{start: 0, end: len(text) + offset})
			}

			return nil, NewExprErr(
				fmt.Sprintf("unknown token: `%s`", text[loc[0]:loc[1]]),
				Location{
					start: loc[0] + offset,
					end:   loc[1] + offset,
				},
			)
		}

		tValue = text[loc[0]:loc[1]]

		tokens = append(tokens, Token{
			kind: tKind,
			text: tValue,
			loc: Location{
				start: loc[0] + offset,
				end:   loc[1] + offset,
			},
		})

		offset += len(tValue)
		text = text[loc[1]:]

		text, trimmed = utils.TrimWhitespacesAndCount(text)
		offset += trimmed
	}

	return tokens, nil
}
