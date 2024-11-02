package utils

import "fmt"

type Queue[T any] []T

func (s Queue[T]) IsEmpty() bool {
	return len(s) == 0
}

func (s Queue[T]) Enqueue(v T) Queue[T] {
	return append(s, v)
}

func (s Queue[T]) Dequeue() (Queue[T], T, error) {
	if len(s) == 0 {
		var zeroValue T
		return s, zeroValue, fmt.Errorf("очередь пуста")
	}
	return s[1:], s[0], nil
}