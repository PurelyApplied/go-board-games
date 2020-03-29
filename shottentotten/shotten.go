package shottentotten

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
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

type cardSet []clanCard

type stone [2]cardSet

func (s stone) display() string {
	left, right := fmt.Sprintf("%v", s[0]), fmt.Sprintf("%v", s[1])

	return fmt.Sprintf("%15v | %15v", left, right)
}

type battleLine [9]stone

func (l battleLine) display() string {
	var stones []string
	for _, s := range l {
		stones = append(stones, s.display())
	}
	return "Battle line:\n------------\n" + strings.Join(stones, "\n")
}

type beginTurnInstruction bool

type playCard struct {
	card clanCard
	loc  int
}

func player(id int, pc playerChan, g gameChan) {
	var hand []clanCard
	for {
		select {
		case in := <-pc:
			switch v := in.(type) {
			case drawInstruction:
				log.Printf("Player %d draws %v\n", id, v)
				hand = append(hand, clanCard(v))
			case beginTurnInstruction:
				log.Printf("Player %d to play...\n", id)
				// TODO state := examineGameState()
				// TODO let's play with strategy

				// randomly select a card and a destination
				i := rand.Intn(len(hand))
				loc := rand.Intn(9)
				card := hand[i]
				// TODO probably out of bounds?
				hand = append(hand[:i], hand[i+1:]...)
				toPlay := playCard{card, loc}
				log.Printf("Player %d to play: %v\n", id, toPlay)
				g <- toPlay
			}
		}
	}
}

type playerChan chan interface{}
type gameChan chan interface{}

func Main(seed int64) {
	rand.Seed(seed)

	chans := []playerChan{make(playerChan), make(playerChan)}
	respChans := []gameChan{make(gameChan), make(gameChan)}
	go player(0, chans[0], respChans[0])
	go player(1, chans[1], respChans[1])

	manageGames(chans, respChans)
}

func shuffleDeck(deck clanDeck) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}

type drawInstruction clanCard

func manageGames(chans []playerChan, respChans []gameChan) {

	deck := newClanDeck()
	shuffleDeck(deck)

	var line battleLine

	dealStartingHands(6, deck, chans)

	officiateGame(deck, line, chans, respChans)
}

func officiateGame(deck clanDeck, line battleLine, chans []playerChan, respChans []gameChan) {
	log.Print("Begin!")
	chans[0] <- beginTurnInstruction(true)
	for {
		for id := 0; id < 2; id++ {
			in := respChans[id]
			out := chans[id]
			select {
			case input := <-in:
				switch instr := input.(type) {
				case playCard:
					log.Printf("Got card %v for position %d from player %d\n", instr.card, instr.loc, id)
					stoneSet := line[instr.loc][id]
					line[instr.loc][id] = append(stoneSet, instr.card)

					// send card
					var draw clanCard
					draw, deck = deck[len(deck)-1], deck[:len(deck)-1]
					log.Printf("Player %v draws %v\n", id, draw)
					out <- drawInstruction(draw)

					// switch player
					other := 1 - id
					log.Printf("Instructing %v to being their turn\n", other)
					chans[other] <- beginTurnInstruction(true)

				default:
					log.Printf("Got %v (%T) from player %d", input, input, id)
				}

			default:
				log.Printf("Nothing from player %d...\n", id)
				time.Sleep(500 * time.Millisecond)
				fmt.Println(line.display())
			}
		}
	}

	// TODO officiate
}

func dealStartingHands(n int, deck clanDeck, chans []playerChan) {
	for i := 0; i < 6; i++ {
		for id := 0; id < 2; id++ {
			var card clanCard
			card, deck = deck[len(deck)-1], deck[:len(deck)-1]

			log.Printf("Player %v draws a %v\n", id, card)
			chans[id] <- drawInstruction(card)
		}
	}
}
