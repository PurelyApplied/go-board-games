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
	rumor       chan rumorMsg
	accuse      chan accuseMsg
	move        chan moveMsg
	startTurn   chan startTurnMsg
	endTurn     chan endTurnMsg
	beDealtItem chan items.Item
	beShownItem chan items.Item
}

func newPlayerComm() playerComm {
	return playerComm{
		rumor:       make(chan rumorMsg),
		accuse:      make(chan accuseMsg),
		move:        make(chan moveMsg),
		startTurn:   make(chan startTurnMsg),
		endTurn:     make(chan endTurnMsg),
		beDealtItem: make(chan items.Item),
		beShownItem: make(chan items.Item),
	}
}

func Clue(nPlayers int) {
	c := make([]playerComm, 0, nPlayers)
	for i := 0; i < nPlayers; i++ {
		c = append(c, newPlayerComm())
	}

	itemSet := items.NewItemSet(6, 6, 10, time.Now().UnixNano())
	actual, deck := itemSet.Setup()

	for id := 0; id < nPlayers; id++ {
		go play(id, c, itemSet)
	}

	// A deck of size N divides among p players N // p times, up to (N // p) * p.  Deal cards.
	for i := 0; i < len(deck)/nPlayers*nPlayers; i++ {
		c[i%nPlayers].beDealtItem <- deck[i]
	}

	// everyone gets to see cards that don't deal out evenly
	for i := len(deck) / nPlayers * nPlayers; i < len(deck); i++ {
		for j := 0; j < nPlayers; j++ {
			c[j].beShownItem <- deck[i]
		}
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
	var myItems []items.Item
	for {
		select {
		case item := <-c[id].beDealtItem:
			log.Printf("Player %d is dealt: %v\n", id, item)
			myItems = append(myItems, item)
		case item := <-c[id].beShownItem:
			log.Printf("Player %d is shown: %v\n", id, item)
			// TODO Think about that.
		case <-c[id].startTurn:
			// TODO moving to rooms
			g := guessRandomly(itemSet)
			c[id].rumor <- rumorMsg(g)
			// TODO get responses
			c[id].endTurn <- "end"

		default:
			log.Printf("Player %d snoozes...\n", id)
			time.Sleep(10 * time.Millisecond)
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
