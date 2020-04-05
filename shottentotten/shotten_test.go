package shottentotten

import (
	"fmt"
	"go-board-games/shottentotten/data/battleline"
	"go-board-games/shottentotten/data/deck"
	"testing"
)

func TestFoo(t *testing.T) {
	line := battleline.newBattleline()
	line.appendTo(1, 1, deck.ClanCard{2, "b"})

	cpy := line.get()
	cpy[0][0] = battleline.CardSet{deck.ClanCard{1, "a"}}

	fmt.Println(line.line)
	fmt.Println(cpy)

}
