package clue

import "testing"

func TestFoo(t *testing.T) {
	n := 10
	p := 3
	t.Log(n / p)
	t.Log(n / p * p)
}
