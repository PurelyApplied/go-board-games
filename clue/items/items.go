package items

import "math/rand"

type item interface{ isItem() }

type Suspect string

func (Suspect) isItem() {}

var suspects = []item{
	// Classic Clue
	Suspect("Mrs. White"),
	Suspect("Mr. Green"),
	Suspect("Mrs. Peacock"),
	Suspect("Professor Plum"),
	Suspect("Miss Scarlett"),
	Suspect("Colonel Mustard"),
	// Just for fun
	Suspect("Mr. Man"),
	Suspect("Dr. Dude"),
	Suspect("Ms. Anthropy"),
	Suspect("Agent Lee Gently"),
	Suspect("Mrs. Ippy"),
	Suspect("Sir Nightly"),
}

type Weapon string

func (Weapon) isItem() {}

var weapons = []item{
	// Classic Clue
	Weapon("candlestick"),
	Weapon("knife"),
	Weapon("lead pipe"),
	Weapon("revolver"),
	Weapon("rope"),
	Weapon("wrench"),
	// Later editions
	Weapon("poison"),
	Weapon("chalice"),
	Weapon("trophy"),
	Weapon("axe"),
	Weapon("dumbbell"),
	Weapon("baseball bat"),
	Weapon("horeshoe"),
	Weapon("hammer"),
	Weapon("garden shears"),
	Weapon("water bucket"),
	Weapon("tennis racquet"),
	Weapon("lawn gnome"),
	// Just for fun
	Weapon("baleful eye"),
	Weapon("cutting remark"),
}

type Location string

func (Location) isItem() {}

var locations = []item{
	// Classic Clue
	Location("ballroom"),
	Location("billiard room"),
	Location("cellar"),
	Location("conservatory"),
	Location("dining room"),
	Location("hall"),
	Location("kitchen"),
	Location("library"),
	Location("lounge"),
	Location("study"),
	// Later revisions
	Location("carriage house"),
	Location("courtyard"),
	Location("drawing room"),
	Location("garage"),
	Location("gazebo."),
	Location("guest House"),
	Location("master bedroom"),
	Location("observatory"),
	Location("patio"),
	Location("pool"),
	Location("spa"),
	Location("studio"),
	Location("theater"),
	Location("trophy room"),
	// Just for fun
	Location("blink of an eye"),
	Location("butt"),
	Location("feels"),
	Location("kisser"),
	Location("nick of time"),
	Location("sense of self"),
}

func ClassicItemSets() ([]Suspect, []Weapon, []Location) {
	s := append([]Suspect{}, toS(toI(weapons[:6]))...)
	w := append([]Weapon{}, toW(toI(weapons[:6]))...)
	l := append([]Location{}, toL(toI(weapons[:10]))...)
	return s, w, l

}
func NewItemSets(s, w, l int, seed int64) ([]Suspect, []Weapon, []Location) {
	// func NewItemSets(w, l, s int, seed int64)
	rand.Seed(seed)
	return toS(choose(s, toI(suspects))), toW(choose(w, toI(weapons))), toL(choose(l, toI(locations)))
}

// SooOOoo reusable!
func toI(items []item) []interface{} {
	ins := make([]interface{}, len(items))
	for i, x := range items {
		ins[i] = x
	}
	return ins
}

func toW(items []interface{}) []Weapon {
	ins := make([]Weapon, len(items))
	for i, x := range items {
		ins[i] = x.(Weapon)
	}
	return ins
}

func toL(items []interface{}) []Location {
	ins := make([]Location, len(items))
	for i, x := range items {
		ins[i] = x.(Location)
	}
	return ins
}

func toS(items []interface{}) []Suspect {
	ins := make([]Suspect, len(items))
	for i, x := range items {
		ins[i] = x.(Suspect)
	}
	return ins
}

func choose(n int, s []interface{}) []interface{} {
	ss := append([]interface{}{}, s...) // for the sake of argument, suppose a shallow copy is okay
	rand.Shuffle(len(s), func(i, j int) {
		ss[i], ss[j] = ss[j], ss[i]
	})
	return ss[:n]
}
