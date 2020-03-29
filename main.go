package main

import (
	"go-board-games/shottentotten"
	"time"
)

//import "go-board-games/clue"

func main() {
	//clue.Clue(4)

	shottentotten.Main(time.Now().UnixNano())
}
