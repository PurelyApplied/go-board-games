package shottentotten

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

const updateInterval = 10 * time.Millisecond

var clans = []string{"a", "b", "c", "d", "e", "f"}

type clanCard struct {
	rank int
	clan string // "suit"
}

func (c clanCard) String() string {
	return fmt.Sprintf("%d%s", c.rank, c.clan)
}

type clanDeck struct {
	cards []clanCard
	sync.RWMutex
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
	}
}

func (cd *clanDeck) shuffle() {
	cd.RWMutex.Lock()
	defer cd.RWMutex.Unlock()

	rand.Shuffle(len(cd.cards), func(i, j int) {
		cd.cards[i], cd.cards[j] = cd.cards[j], cd.cards[i]
	})
}

func (cd *clanDeck) draw() (draw clanCard, ok bool) {
	cd.RWMutex.Lock()
	defer cd.RWMutex.Unlock()

	if len(cd.cards) == 0 {
		return clanCard{}, false
	}

	log.Printf("Drawing from a deck with %d cards left...\n", len(cd.cards))
	draw, cd.cards = cd.cards[len(cd.cards)-1], cd.cards[:len(cd.cards)-1]
	return draw, true
}

func (cd *clanDeck) size() int {
	cd.RWMutex.RLock()
	defer cd.RWMutex.RUnlock()

	return len(cd.cards)
}

type cardSet struct {
	cards []clanCard
	sync.RWMutex
}

func (s *cardSet) isFlush() bool {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	colors := make(map[string]bool)
	for _, c := range s.cards {
		colors[c.clan] = true
	}
	return len(colors) == 1
}

func (s *cardSet) isRun() bool {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	bins := make([]int, 9, 9)
	for _, c := range s.cards {
		bins[c.rank] += 1
	}

	counter := 0
	for i := 0; i < 9; i++ {
		if bins[i] > 1 {
			return false
		}
		if bins[i] == 1 {
			counter++
		} else {
			counter = 0
		}

		if counter == 3 {
			return true
		}
	}
	return false
}

func (s *cardSet) highCard() int {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	high := -1
	for _, v := range s.cards {
		if v.rank > high {
			high = v.rank
		}
	}

	return high
}

func (s *cardSet) isTriple() bool {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	bins := make([]int, 9, 9)
	for _, c := range s.cards {
		bins[c.rank] += 1
	}

	for i := 0; i < 9; i++ {
		if bins[i] == 3 {
			return true
		}
	}
	return false
}

func (s *cardSet) sum() int {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	total := 0
	for _, v := range s.cards {
		total += v.rank
	}
	return total
}

func newCardSet() *cardSet {
	return &cardSet{
		cards: make([]clanCard, 0, 3),
	}
}

type winner int

const (
	tbd   winner = -1
	left  winner = 0
	right winner = 1
)

type stone struct {
	// TODO Needs history as tie-breaker for, e.g., 3 6s vs 3 6s
	cards [2]*cardSet
	winner
	sync.RWMutex
}

func newStone() *stone {
	s := &stone{winner: tbd}
	s.cards[0] = newCardSet()
	s.cards[1] = newCardSet()
	return s
}

func (s *stone) Display() string {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	separator := s.getWinnerStr()

	left, right := fmt.Sprintf("%v", s.cards[0].cards), fmt.Sprintf("%v", s.cards[1].cards)
	return fmt.Sprintf("%15v%s%-15v", left, separator, right)
}

func (s *stone) getWinnerStr() interface{} {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	switch s.winner {
	case tbd:
		return " | "
	case left:
		return " | "
	case right:
		return " | "
	default:
		log.Printf("E] Got unexpected winner at stone.")
		return "?|?"
	}
}

type handKind int

const (
	// TODO rename
	sum handKind = iota
	run
	color
	three
	colorRun
)

type strength struct {
	handKind
	value int
}

func evaluateCards(set *cardSet) strength {
	flush := set.isFlush()
	isRun := set.isRun()
	triple := set.isTriple()
	highCard := set.highCard()

	switch {
	case flush && isRun:
		return strength{handKind: colorRun, value: highCard}
	case triple:
		return strength{handKind: three, value: highCard}
	case flush:
		return strength{handKind: color, value: highCard}
	case isRun:
		return strength{handKind: run, value: highCard}
	default:
		return strength{handKind: sum, value: set.sum()}
	}
}

func (s *stone) updateWinner() {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	if s.winner != tbd {
		return
	}

	if len(s.cards[0].cards) < 3 || len(s.cards[1].cards) < 3 {
		return
	}

	l, r := evaluateCards(s.cards[0]), evaluateCards(s.cards[1])

	switch {
	case l.handKind > r.handKind:
		s.winner = left
	case l.handKind < r.handKind:
		s.winner = right
	case l.value > r.value:
		s.winner = left
	case l.value < r.value:
		s.winner = right
	}
}

type battleLine struct {
	line []*stone
	sync.RWMutex
}

func (l *battleLine) String() string {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()
	return fmt.Sprintf("%v", l.line)
}

func (l *battleLine) display() string {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()

	var stones []string
	for _, s := range l.line {
		stones = append(stones, s.Display())
	}
	return "Battle line:\n------------\n" + strings.Join(stones, "\n")
}

func (l *battleLine) get() []*stone {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()

	cpy := make([]*stone, 9, 9)
	copy(cpy, l.line)
	return cpy
}

func (l *battleLine) appendTo(i, side int, c clanCard) {
	l.RWMutex.Lock()
	defer l.RWMutex.Unlock()

	l.line[i].cards[side].cards = append(l.line[i].cards[side].cards, c)
}

func (l *battleLine) updateStoneWinners() {
	l.RWMutex.Lock()
	defer l.RWMutex.Unlock()

	for _, s := range l.line {
		s.updateWinner()
	}
}

func newBattleline() *battleLine {
	line := battleLine{
		line: make([]*stone, 9, 9),
	}
	for i := 0; i < 9; i++ {
		line.line[i] = newStone()
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
				linedata := line.get()
				var openings []int
				for i, stone := range linedata {
					side := stone.cards[id].cards
					if len(side) < 3 {
						openings = append(openings, i)
					}
				}

				iLoc := rand.Intn(len(openings))
				loc := openings[iLoc]

				i := rand.Intn(len(hand))
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
	go server(deck, line)
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
				time.Sleep(updateInterval)
				fmt.Println(line.display())
				line.updateStoneWinners()

			}
		}
	}
}

var templates = template.Must(template.ParseFiles("shottentotten/game-view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, line *battleLine) {
	err := templates.ExecuteTemplate(w, tmpl+".html", struct{ Display string }{line.display()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func server(deck *clanDeck, line *battleLine) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "game-view", line)
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
