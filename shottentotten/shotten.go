package shottentotten

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var clans = []string{"a", "b", "c", "d", "e", "f"}

type clanCard struct {
	rank int
	clan string // "suit"
}

type clanDeck struct {
	cards []clanCard
	lock  sync.Mutex
}

func newClanDeck() *clanDeck {
	var cards []clanCard
	for _, c := range clans {
		for r := 1; r <= 9; r++ {
			cards = append(cards, clanCard{
				rank: r,
				clan: c,
			})
		}
	}

	return &clanDeck{
		cards: cards,
		lock:  sync.Mutex{},
	}
}

func (cd *clanDeck) shuffle() {
	cd.lock.Lock()
	defer cd.lock.Unlock()

	rand.Shuffle(len(cd.cards), func(i, j int) {
		cd.cards[i], cd.cards[j] = cd.cards[j], cd.cards[i]
	})
}

func (cd *clanDeck) draw() (draw clanCard, ok bool) {
	cd.lock.Lock()
	defer cd.lock.Unlock()

	if len(cd.cards) == 0 {
		return clanCard{}, false
	}

	log.Printf("Drawing from a deck with %d cards left...\n", len(cd.cards))
	draw, cd.cards = cd.cards[len(cd.cards)-1], cd.cards[:len(cd.cards)-1]
	return draw, true
}

type cardSet []clanCard

func displayStone(set [2]cardSet) string {
	left, right := fmt.Sprintf("%v", set[0]), fmt.Sprintf("%v", set[1])
	return fmt.Sprintf("%60v | %-60v", left, right)
}

type battleLine struct {
	line [][2]cardSet
	lock sync.Mutex
}

func (l *battleLine) display() string {
	l.lock.Lock()
	defer l.lock.Unlock()

	var stones []string
	for _, s := range l.line {
		stones = append(stones, displayStone(s))
	}
	return "Battle line:\n------------\n" + strings.Join(stones, "\n")
}

func (l *battleLine) get() [][2]cardSet {
	l.lock.Lock()
	defer l.lock.Unlock()

	var cpy [][2]cardSet
	copy(cpy, l.line)
	return cpy
}

func (l *battleLine) appendTo(i, side int, c clanCard) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.line[i][side] = append(l.line[i][side], c)
}

func newBattleline() *battleLine {
	line := battleLine{
		line: make([][2]cardSet, 9, 9),
		lock: sync.Mutex{},
	}
	return &line
}

type playerInstructionBeginTurn struct{}
type playerInstructionDrawCard struct{}

type playCard struct {
	card clanCard
	loc  int
}

func player(id int, chans chanGroup, deck *clanDeck, line *battleLine) {
	var hand []clanCard
	for {
		select {
		case in := <-chans.toPlayer:
			switch in.(type) {
			case playerInstructionDrawCard:
				draw, ok := deck.draw()
				if !ok {
					log.Printf("No cards for Player %d to draw!\n", id)
				} else {
					log.Printf("Player %d draws %v\n", id, draw)
					hand = append(hand, draw)
				}
			case playerInstructionBeginTurn:
				log.Printf("Player %d to play...\n", id)
				if len(hand) == 0 {
					log.Printf("Player %d has no cards to play!!\n", id)
					break
				}

				// randomly select a card and a destination
				i := rand.Intn(len(hand))
				loc := rand.Intn(9)
				card := hand[i]
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

func officiateGame(deck *clanDeck, line *battleLine, chans []chanGroup) {
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

					line.appendTo(instr.loc, id, instr.card)

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
