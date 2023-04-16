package executor

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mymmrac/mm/utils"
)

func init() {
	opText := utils.Keys(textToOp)

	var singleCharOps, multiCharOps []string
	utils.ForeachSlice(opText, func(op string) {
		if len(op) == 1 {
			singleCharOps = append(singleCharOps, op)
		} else {
			multiCharOps = append(multiCharOps, op)
		}
	})

	escapeAll := func(op string) string {
		escaped := ""
		for _, r := range op {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' {
				escaped += fmt.Sprintf("%c", r)
				continue
			}
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

type Lexer struct{}

func NewLexer() *Lexer {
	return &Lexer{}
}

var (
	identifierPattern = regexp.MustCompile(`^[a-zA-Z]\w*`)
	numberPattern     = regexp.MustCompile(`^[\d_]+(:?\.[\d_]+)?`) // TODO: Fix cases like _1
	operatorPattern   *regexp.Regexp

	// TODO: Check if valid
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
		loc = numberPattern.FindStringIndex(text)
		tKind = KindNumber

		if len(loc) == 0 {
			loc = identifierPattern.FindStringIndex(text)
			tKind = KindIdentifier
		}

		if len(loc) == 0 {
			loc = operatorPattern.FindStringIndex(text)
			tKind = KindOperator
		}

		if len(loc) == 0 {
			loc = unknownPattern.FindStringIndex(text)

			if len(loc) == 0 {
				// TODO: Check if loc is correct
				return nil, NewExprErr("invalid expression", Location{Start: 0, End: len(text) + offset})
			}

			return nil, NewExprErr(
				fmt.Sprintf("unknown token: `%s`", text[loc[0]:loc[1]]),
				Location{
					Start: loc[0] + offset,
					End:   loc[1] + offset,
				},
			)
		}

		tValue = text[loc[0]:loc[1]]

		tokens = append(tokens, Token{
			kind: tKind,
			text: tValue,
			loc: Location{
				Start: loc[0] + offset,
				End:   loc[1] + offset,
			},
		})

		offset += len(tValue)
		text = text[loc[1]:]

		text, trimmed = utils.TrimWhitespacesAndCount(text)
		offset += trimmed
	}

	return tokens, nil
}
