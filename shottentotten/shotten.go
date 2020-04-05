package shottentotten

import (
	"fmt"
	"go-board-games/shottentotten/data/deck"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const updateInterval = 10 * time.Millisecond

type playerInstructionBeginTurn struct{}
type playerInstructionDrawCard struct{}

type playCard struct {
	card deck.ClanCard
	loc  int
}

func player(id int, chans chanGroup, dk *deck.ClanDeck, line *battleLine) {
	var hand []deck.ClanCard
	for {
		select {
		case in := <-chans.toPlayer:
			switch in.(type) {
			case playerInstructionDrawCard:
				draw, ok := dk.Draw()
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

	dk := deck.New()
	line := newBattleline()
	chans := []chanGroup{newChanGroup(0), newChanGroup(1)}

	go player(0, chans[0], dk, line)
	go player(1, chans[1], dk, line)

	officiateGame(dk, line, chans)
}

type historicMove struct {
	id   int
	card deck.ClanCard
}

func officiateGame(deck *deck.ClanDeck, line *battleLine, chans []chanGroup) {
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

func server(dk *deck.ClanDeck, line *battleLine) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "game-view", line)
	}

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
