package main

import "fmt"

func mapSlice[T1, T2 any](slice []T1, f func(T1) T2) []T2 {
	newSlice := make([]T2, len(slice))
	for i, v := range slice {
		newSlice[i] = f(v)
	}
	return newSlice
}

func toString[T fmt.Stringer](a T) string {
	return a.String()
}
