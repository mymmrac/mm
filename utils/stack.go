package utils

import "fmt"

type Stack[T any] struct {
	values []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Push(a ...T) {
	s.values = append(s.values, a...)
}

func (s *Stack[T]) Pop() T {
	value := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return value
}

func (s *Stack[T]) Top() T {
	return s.values[len(s.values)-1]
}

func (s *Stack[T]) Empty() bool {
	return len(s.values) == 0
}

func (s *Stack[T]) Size() int {
	return len(s.values)
}

func (s *Stack[T]) Slice() []T {
	return s.values
}

func (s *Stack[T]) String() string {
	return fmt.Sprint(s.values)
}
