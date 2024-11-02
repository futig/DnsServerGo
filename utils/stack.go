package utils

import "fmt"

type Stack[T any] []T

func (s Stack[T]) Push(v T) Stack[T] {
	return append(s, v)
}

func (s Stack[T]) IsEmpty() bool {
	return len(s) == 0
}

func (s Stack[T]) PushRange(v []T) Stack[T] {
	return append(s, v...)
}

func (s Stack[T]) Pop() (Stack[T], T, error) {
	if len(s) == 0 {
		var zeroValue T
		return s, zeroValue, fmt.Errorf("стек пуст")
	}
	l := len(s)
	return s[:l-1], s[l-1], nil
}
