package utils

type Stack[T any] struct {
	values []T
}

func (s *Stack[T]) Push(a T) {
	s.values = append(s.values, a)
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
