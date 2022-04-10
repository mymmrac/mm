package utils

import (
	"fmt"
	"strings"
)

func MapSlice[T1, T2 any](slice []T1, f func(T1) T2) []T2 {
	newSlice := make([]T2, len(slice))
	for i, v := range slice {
		newSlice[i] = f(v)
	}
	return newSlice
}

func ForeachSlice[T any](slice []T, f func(T)) {
	for _, v := range slice {
		f(v)
	}
}

func ToString[T fmt.Stringer](a T) string {
	return a.String()
}

func TrimWhitespacesAndCount(text string) (string, int) {
	newText := strings.TrimLeft(text, " \t")
	return newText, len(text) - len(newText)
}

func Assert(ok bool, args ...any) {
	if !ok {
		panic(fmt.Sprint(args...))
	}
}

func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}