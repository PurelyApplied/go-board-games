package shottentotten

import (
	"fmt"
	"testing"
)

func TestFoo(t *testing.T) {
	line := newBattleline()
	line.appendTo(1, 1, clanCard{2, "b"})

	cpy := line.get()
	cpy[0][0] = cardSet{clanCard{1, "a"}}

	fmt.Println(line.line)
	fmt.Println(cpy)

}
