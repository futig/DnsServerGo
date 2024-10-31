package utils

import "fmt"

type Stack []string

func (s Stack) Push(v string) Stack {
	return append(s, v)
}

func (s Stack) IsEmpty() bool {
	return len(s) == 0
}

func (s Stack) PushRange(v []string) Stack {
	return append(s, v...)
}

func (s Stack) Pop() (Stack, string, error) {
	if len(s) == 0 {
		return s, "", fmt.Errorf("Стек пуст")
	}
	l := len(s)
	return s[:l-1], s[l-1], nil
}
