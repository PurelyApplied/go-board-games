package shottentotten

import (
	"fmt"
	"go-board-games/shottentotten/data/deck"
	"testing"
)

func TestFoo(t *testing.T) {
	line := newBattleline()
	line.appendTo(1, 1, deck.ClanCard{2, "b"})

	cpy := line.get()
	cpy[0][0] = cardSet{deck.ClanCard{1, "a"}}

	fmt.Println(line.line)
	fmt.Println(cpy)

}
