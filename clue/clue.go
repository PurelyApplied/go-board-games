package clue

import (
	"log"
	"math/rand"
	"time"
)

type rumorMsg interface{}
type accuseMsg interface{}
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

	for id := 0; id < nPlayers; id++ {
		go play(id, c)
	}

	main(nPlayers, c)
}

func main(nPlayers int, c []playerComm) {
	for currentPlayer := rand.Intn(nPlayers); ; currentPlayer = (currentPlayer + 1) % nPlayers {
		c[currentPlayer].startTurn <- "start"
		<-c[currentPlayer].endTurn

		time.Sleep(500 * time.Millisecond)
	}
}

func play(id int, c []playerComm) {
	for {
		select {
		case <-c[id].startTurn:
			log.Println("Player", id, "takes their turn...")
			c[id].endTurn <- "end"
		}
	}
}

type guessTactic string
type retention string

const (
	guessRandom   guessTactic = "guess-random"
	guessInformed guessTactic = "guess-informed"

	retainNothing retention = "nothing"
	retainShown   retention = "shown"
	retainPasses  retention = "shown"
)
