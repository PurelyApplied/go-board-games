package shottentotten

import (
	"fmt"
	"go-board-games/shottentotten/data/deck"
	"log"
	"strings"
	"sync"
)

type CardSet struct {
	cards []deck.ClanCard
	sync.RWMutex
}

func (s *CardSet) isFlush() bool {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	colors := make(map[string]bool)
	for _, c := range s.cards {
		colors[c.Clan] = true
	}
	return len(colors) == 1
}

func (s *CardSet) isRun() bool {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	bins := make([]int, 9, 9)
	for _, c := range s.cards {
		bins[c.Rank] += 1
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

func (s *CardSet) highCard() int {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	high := -1
	for _, v := range s.cards {
		if v.Rank > high {
			high = v.Rank
		}
	}

	return high
}

func (s *CardSet) isTriple() bool {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	bins := make([]int, 9, 9)
	for _, c := range s.cards {
		bins[c.Rank] += 1
	}

	for i := 0; i < 9; i++ {
		if bins[i] == 3 {
			return true
		}
	}
	return false
}

func (s *CardSet) sum() int {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	total := 0
	for _, v := range s.cards {
		total += v.Rank
	}
	return total
}

func newCardSet() *CardSet {
	return &CardSet{
		cards: make([]deck.ClanCard, 0, 3),
	}
}

type winner int

const (
	tbd   winner = -1
	left  winner = 0
	right winner = 1
)

type Stone struct {
	// TODO Needs history as tie-breaker for, e.g., 3 6s vs 3 6s
	cards [2]*CardSet
	winner
	sync.RWMutex
}

func newStone() *Stone {
	s := &Stone{winner: tbd}
	s.cards[0] = newCardSet()
	s.cards[1] = newCardSet()
	return s
}

func (s *Stone) Display() string {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	separator := s.getWinnerStr()

	left, right := fmt.Sprintf("%v", s.cards[0].cards), fmt.Sprintf("%v", s.cards[1].cards)
	return fmt.Sprintf("%15v%s%-15v", left, separator, right)
}

func (s *Stone) getWinnerStr() interface{} {
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

type HandKind int

const (
	sumHK HandKind = iota
	runHK
	colorHK
	tripleHK
	colorRunHK
)

type Strength struct {
	HandKind
	Value int
}

func EvaluateCards(set *CardSet) Strength {
	flush := set.isFlush()
	isRun := set.isRun()
	triple := set.isTriple()
	highCard := set.highCard()

	switch {
	case flush && isRun:
		return Strength{HandKind: colorRunHK, Value: highCard}
	case triple:
		return Strength{HandKind: tripleHK, Value: highCard}
	case flush:
		return Strength{HandKind: colorHK, Value: highCard}
	case isRun:
		return Strength{HandKind: runHK, Value: highCard}
	default:
		return Strength{HandKind: sumHK, Value: set.sum()}
	}
}

func (s *Stone) updateWinner() {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	if s.winner != tbd {
		return
	}

	if len(s.cards[0].cards) < 3 || len(s.cards[1].cards) < 3 {
		return
	}

	l, r := EvaluateCards(s.cards[0]), EvaluateCards(s.cards[1])

	switch {
	case l.HandKind > r.HandKind:
		s.winner = left
	case l.HandKind < r.HandKind:
		s.winner = right
	case l.Value > r.Value:
		s.winner = left
	case l.Value < r.Value:
		s.winner = right
	}
}

type battleLine struct {
	line []*Stone
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

func (l *battleLine) get() []*Stone {
	l.RWMutex.RLock()
	defer l.RWMutex.RUnlock()

	cpy := make([]*Stone, 9, 9)
	copy(cpy, l.line)
	return cpy
}

func (l *battleLine) appendTo(i, side int, c deck.ClanCard) {
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
		line: make([]*Stone, 9, 9),
	}
	for i := 0; i < 9; i++ {
		line.line[i] = newStone()
	}
	return &line
}
