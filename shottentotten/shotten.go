package shottentotten

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

var clans = []string{"a", "b", "c", "d", "e", "f"}

type clanCard struct {
	rank int
	clan string // "suit"
}

type copyable interface {
	copy() copyable
}

type querable struct {
	value   copyable
	request chan struct{}
	respond chan interface{}
}

func (q *querable) makeChans() {
	q.request = make(chan struct{})
	q.respond = make(chan interface{})
}

func (q querable) listen() {
	for {
		select {
		case <-q.request:
			q.respond <- q.value.copy()
		}
	}
}

type clanDeck querable

type clanCardSlice []clanCard

func (ccs clanCardSlice) copy() copyable {
	var cpy clanCardSlice
	copy(cpy, ccs)
	return cpy
}

func newClanDeck() clanDeck {

	deck := clanDeck{
		value:   nil,
		request: make(chan struct{}),
		respond: make(chan interface{}),
	}

	defer func() { go deck.listen() }()

	var cards clanCardSlice
	for _, c := range clans {
		for r := 1; r <= 9; r++ {
			cards = append(cards, clanCard{
				rank: r,
				clan: c,
			})
		}
	}

	deck.value = cards
	return deck
}

func (cd *clanDeck) shuffle() {
	cards := cd.value.(clanCardSlice)
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}

func (cd *clanDeck) draw() clanCard {
	cd.request <- struct{}{}
	return (<-cd.respond).(clanCard)
}

// deck listens for draws
func (cd *clanDeck) listen() {
	for {
		select {
		case <-cd.request:
			deck := cd.value.(clanCardSlice)
			draw, deck := deck[len(deck)-1], deck[:len(deck)-1]
			cd.value = deck
			cd.respond <- draw
		}
	}

}

type cardSet []clanCard

type stone [2]cardSet

func (s stone) display() string {
	left, right := fmt.Sprintf("%v", s[0]), fmt.Sprintf("%v", s[1])

	return fmt.Sprintf("%60v | %-60v", left, right)
}

type battleLine struct {
	line     [9]stone
	reqChan  lineChan
	respChan chan [9]stone
}

func (l battleLine) display() string {
	var stones []string
	for _, s := range l.line {
		stones = append(stones, s.display())
	}
	return "Battle line:\n------------\n" + strings.Join(stones, "\n")
}

func (l battleLine) get() [9]stone {
	l.reqChan <- ""
	return <-l.respChan
}

func (l battleLine) listen() {
	for {
		select {
		case <-l.reqChan:
			l.respChan <- l.copy()
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (l battleLine) copy() [9]stone {
	cpy := [9]stone{}
	for i := 0; i < 9; i++ {
		cpy[i] = l.line[i]
	}
	return cpy
}

func newBattleline() *battleLine {
	line := battleLine{
		line:     [9]stone{},
		reqChan:  make(lineChan),
		respChan: make(chan [9]stone),
	}
	go line.listen()
	return &line
}

type playerInstructionBeginTurn struct{}
type playerInstructionDrawCard struct{}

type playCard struct {
	card clanCard
	loc  int
}

func player(id int, chans chanGroup, deck clanDeck, line *battleLine) {
	var hand []clanCard
	for {
		select {
		case in := <-chans.toPlayer:
			switch in.(type) {
			case playerInstructionDrawCard:
				draw := deck.draw()
				log.Printf("Player %d draws %v\n", id, draw)
				hand = append(hand, draw)
			case playerInstructionBeginTurn:
				log.Printf("Player %d to play...\n", id)
				_ = line
				// randomly select a card and a destination
				i := rand.Intn(len(hand))
				loc := rand.Intn(9)
				card := hand[i]
				// TODO probably out of bounds?
				hand = append(hand[:i], hand[i+1:]...)
				toPlay := playCard{card, loc}
				log.Printf("Player %d to play: %v\n", id, toPlay)
				chans.toGame <- toPlay
			}
		}
	}
}

type playerChan chan interface{}
type gameChan chan interface{}
type lineChan chan interface{}
type deckChan chan interface{}

type chanGroup struct {
	id       int
	toPlayer playerChan
	toGame   gameChan
	line     lineChan
	deck     deckChan
}

func newChanGroup(id int) chanGroup {
	return chanGroup{
		id:       id,
		toPlayer: make(playerChan),
		toGame:   make(gameChan),
		line:     make(lineChan),
		deck:     make(deckChan),
	}
}

func Main(seed int64) {
	rand.Seed(seed)

	deck := newClanDeck()
	deck.shuffle()
	line := newBattleline()
	chans := []chanGroup{newChanGroup(0), newChanGroup(1)}

	go player(0, chans[0], deck, line)
	go player(1, chans[1], deck, line)

	officiateGame(deck, line, chans)
}

type historicMove struct {
	id   int
	card clanCard
}

func officiateGame(deck clanDeck, line *battleLine, chans []chanGroup) {
	log.Print("Begin!")
	for i := 0; i < 6; i++ {
		chans[0].toPlayer <- playerInstructionDrawCard{}
		chans[1].toPlayer <- playerInstructionDrawCard{}
	}

	var history []historicMove

	chans[0].toPlayer <- playerInstructionBeginTurn{}
	for {
		for id := 0; id < 2; id++ {
			select {
			case input := <-chans[id].toGame:
				switch instr := input.(type) {
				case playCard:
					log.Printf("Got card %v for position %d from player %d\n", instr.card, instr.loc, id)
					// Add card to history
					history = append(history, historicMove{id, instr.card})

					// Set card to stone
					stoneSet := line.line[instr.loc][id]
					line.line[instr.loc][id] = append(stoneSet, instr.card)

					// instruct to draw
					chans[id].toPlayer <- playerInstructionDrawCard{}

					// switch player
					other := 1 - id
					log.Printf("Instructing %v to being their turn\n", other)
					chans[other].toPlayer <- playerInstructionBeginTurn{}

				default:
					log.Printf("Got %v (%T) from player %d", input, input, id)
				}

			default:
				time.Sleep(10 * time.Millisecond)
				fmt.Println(line.display())
			}
		}
	}

	// TODO officiate
}
