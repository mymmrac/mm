package utils

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
)

func Assert(ok bool, args ...any) {
	if !ok {
		panic(fmt.Sprint(args...))
	}
}

func Wrap(text string, limit int) string {
	return wrap.String(wordwrap.String(text, limit), limit)
}

func IsSpace(c byte) bool {
	switch c {
	case ' ', '\n', '\t', '\v', '\f', '\r', 0x85, 0xA0:
		return true
	}
	return false
}

func IsDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func IsInCharset(c byte, charset string) bool {
	return strings.IndexByte(charset, c) >= 0
}
