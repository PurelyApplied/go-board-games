package clue

import (
	"go-board-games/clue/items"
	"log"
	"math/rand"
	"time"
)

type rumorMsg items.Jaccuse
type accuseMsg items.Jaccuse

type moveMsg interface{}
type endTurnMsg interface{}
type startTurnMsg interface{}

type playerComm struct {
	rumor     chan rumorMsg
	accuse    chan accuseMsg
	move      chan moveMsg
	endTurn   chan endTurnMsg
	startTurn chan startTurnMsg
}

func newPlayerComm() playerComm {
	return playerComm{
		rumor:     make(chan rumorMsg),
		accuse:    make(chan accuseMsg),
		move:      make(chan moveMsg),
		startTurn: make(chan startTurnMsg),
		endTurn:   make(chan endTurnMsg),
	}
}

func Clue(nPlayers int) {
	var c []playerComm
	for i := 0; i < nPlayers; i++ {
		c = append(c, newPlayerComm())
	}

	itemSet := items.NewItemSet(6, 6, 10, time.Now().UnixNano())
	actual, _ := itemSet.Setup()

	// TODO deal cards (minus actual)
	// TODO public knowledge when it doesn't deal cleanly
	for id := 0; id < nPlayers; id++ {
		go play(id, c, itemSet)
	}

	coordinate(nPlayers, c, actual)
}

func coordinate(nPlayers int, c []playerComm, actual items.Jaccuse) {
	currentPlayer := rand.Intn(nPlayers)
	c[currentPlayer].startTurn <- "start"
	gameOver := false
	for !gameOver {
		select {
		case <-c[currentPlayer].endTurn:
			log.Printf("Player %d ending turn.\n", currentPlayer)
			currentPlayer = (currentPlayer + 1) % nPlayers

			log.Printf("Signaling player %d to start their turn.\n", currentPlayer)
			c[currentPlayer].startTurn <- "start"

		case r := <-c[currentPlayer].rumor:
			log.Printf("Player %d spreads a rumor: %v\n", currentPlayer, r)
			log.Printf("(Actual is: %v\n", actual)

			if items.Jaccuse(r) == actual {
				log.Printf("Player %d rumored correctly!  Ending game.\n", currentPlayer)
				gameOver = true
			}
		}
	}
}

func play(id int, c []playerComm, itemSet items.ItemSet) {
	for {
		select {
		case <-c[id].startTurn:
			// TODO moving to rooms
			g := guessRandomly(itemSet)
			c[id].rumor <- rumorMsg(g)
			// TODO get responses
			c[id].endTurn <- "end"
		}
	}
}

func guessRandomly(set items.ItemSet) items.Jaccuse {
	return items.Jaccuse{
		Suspect:  set.Suspects[rand.Intn(len(set.Suspects))],
		Weapon:   set.Weapons[rand.Intn(len(set.Weapons))],
		Location: set.Locations[rand.Intn(len(set.Locations))],
	}
}

// TODO Future snippets:
type guessTactic string
type retention string

const (
	guessRandom   guessTactic = "guess-random"
	guessInformed guessTactic = "guess-informed"

	retainNothing retention = "nothing"
	retainShown   retention = "shown"
	retainPasses  retention = "shown"
)
