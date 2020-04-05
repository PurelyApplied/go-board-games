package shottentotten

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
)

var clans = []string{"a", "b", "c", "d", "e", "f"}

type ClanCard struct {
	rank int
	clan string // "suit"
}

func (c ClanCard) String() string {
	return fmt.Sprintf("%d%s", c.rank, c.clan)
}

type ClanDeck struct {
	cards []ClanCard
	sync.RWMutex
}

func New() *ClanDeck {
	var cards []ClanCard
	for _, c := range clans {
		for r := 1; r <= 9; r++ {
			cards = append(cards, ClanCard{
				rank: r,
				clan: c,
			})
		}
	}

	deck := &ClanDeck{
		cards: cards,
	}
	deck.shuffle()

	return deck
}

func (cd *ClanDeck) shuffle() {
	cd.RWMutex.Lock()
	defer cd.RWMutex.Unlock()

	rand.Shuffle(len(cd.cards), func(i, j int) {
		cd.cards[i], cd.cards[j] = cd.cards[j], cd.cards[i]
	})
}

func (cd *ClanDeck) Draw() (draw ClanCard, ok bool) {
	cd.RWMutex.Lock()
	defer cd.RWMutex.Unlock()

	if len(cd.cards) == 0 {
		return ClanCard{}, false
	}

	log.Printf("Drawing from a deck with %d cards left...\n", len(cd.cards))
	draw, cd.cards = cd.cards[len(cd.cards)-1], cd.cards[:len(cd.cards)-1]
	return draw, true
}

func (cd *ClanDeck) size() int {
	cd.RWMutex.RLock()
	defer cd.RWMutex.RUnlock()

	return len(cd.cards)
}
