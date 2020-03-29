package shottentotten

import (
	"log"
	"math/rand"
)

type clanCard struct {
	rank int
	clan string // "suit"
}
type clanDeck []clanCard

var clans = []string{"a", "b", "c", "d", "e", "f"}

func newClanDeck() clanDeck {
	var deck clanDeck
	for _, c := range clans {
		for r := 1; r <= 9; r++ {
			deck = append(deck, clanCard{
				rank: r,
				clan: c,
			})
		}
	}
	return deck
}

type stone int

type battleLine [9]stone

func player(id int, pc playerChan) {
	var hand []clanCard
	for {
		select {
		case in := <-pc:
			switch v := in.(type) {
			case drawDirective:
				log.Printf("Player %d draws %v\n", id, v)
				hand = append(hand, clanCard(v))
			}
		}
	}
}

type playerChan chan interface{}

func Main(seed int64) {
	rand.Seed(seed)

	chans := []playerChan{make(playerChan), make(playerChan)}

	go player(0, chans[0])
	go player(1, chans[1])

	manageGames(chans)
}

func shuffleDeck(deck clanDeck) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

type drawDirective clanCard

func manageGames(chans []playerChan) {

	deck := newClanDeck()
	shuffleDeck(deck)

	dealStartingHands(6, deck, chans)

	officiateGame(deck, chans)
}

func officiateGame(deck clanDeck, chans []playerChan) {
	// TODO officiate
}

func dealStartingHands(n int, deck clanDeck, chans []playerChan) {
	for i := 0; i < 6; i++ {
		for id := 0; id < 2; id++ {
			var card clanCard
			card, deck = deck[len(deck)-1], deck[:len(deck)-1]

			log.Printf("Player %v draws a %v\n", id, card)
			chans[id] <- drawDirective(card)
		}
	}
}
