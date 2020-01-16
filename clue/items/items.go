package items

import "math/rand"

type Item string

type Suspect Item

var suspects = []Suspect{
	// Classic Clue
	"Mrs. White",
	"Mr. Green",
	"Mrs. Peacock",
	"Professor Plum",
	"Miss Scarlett",
	"Colonel Mustard",
	// Just for fun
	"Mr. Man",
	"Dr. Dude",
	"Ms. Anthropy",
	"Agent Lee Gently",
	"Mrs. Ippy",
	"Sir Nightly",
}

type Weapon Item

var weapons = []Weapon{
	// Classic Clue
	"candlestick",
	"knife",
	"lead pipe",
	"revolver",
	"rope",
	"wrench",
	// Later editions
	"poison",
	"chalice",
	"trophy",
	"axe",
	"dumbbell",
	"baseball bat",
	"horeshoe",
	"hammer",
	"garden shears",
	"water bucket",
	"tennis racquet",
	"lawn gnome",
	// Just for fun
	"baleful eye",
	"cutting remark",
}

type Location Item

var locations = []Location{
	// Classic Clue
	"ballroom",
	"billiard room",
	"cellar",
	"conservatory",
	"dining room",
	"hall",
	"kitchen",
	"library",
	"lounge",
	"study",
	// Later revisions
	"carriage house",
	"courtyard",
	"drawing room",
	"garage",
	"gazebo.",
	"guest House",
	"master bedroom",
	"observatory",
	"patio",
	"pool",
	"spa",
	"studio",
	"theater",
	"trophy room",
	// Just for fun
	"blink of an eye",
	"butt",
	"feels",
	"kisser",
	"nick of time",
	"sense of self",
}

type Jaccuse struct {
	Suspect
	Weapon
	Location
}

type ItemSet struct {
	Suspects  []Suspect
	Weapons   []Weapon
	Locations []Location
}

func (is ItemSet) Setup() (Jaccuse, []Item) {
	actual := Jaccuse{
		Suspect:  is.Suspects[0],
		Weapon:   is.Weapons[0],
		Location: is.Locations[0],
	}

	deck := make([]Item, 0, len(is.Suspects)+len(is.Weapons)+len(is.Locations)-3)
	for _, s := range is.Suspects[1:] {
		deck = append(deck, Item(s))
	}
	for _, w := range is.Weapons[1:] {
		deck = append(deck, Item(w))
	}
	for _, l := range is.Locations[1:] {
		deck = append(deck, Item(l))
	}
	return actual, deck
}

func ClassicItemSets() ItemSet {
	return ItemSet{
		Suspects:  chooseS(6, suspects[:6]),
		Weapons:   chooseW(6, weapons[:6]),
		Locations: chooseL(10, locations[:10]),
	}

}
func NewItemSet(s, w, l int, seed int64) ItemSet {
	rand.Seed(seed)
	return ItemSet{
		Suspects:  chooseS(s, suspects),
		Weapons:   chooseW(w, weapons),
		Locations: chooseL(l, locations),
	}
}

func chooseS(n int, susps []Suspect) []Suspect {
	ss := append([]Suspect{}, susps...)
	rand.Shuffle(len(ss), func(i, j int) {
		ss[i], ss[j] = ss[j], ss[i]
	})
	return ss[:n]
}

func chooseW(n int, weaps []Weapon) []Weapon {
	ss := append([]Weapon{}, weaps...)
	rand.Shuffle(len(ss), func(i, j int) {
		ss[i], ss[j] = ss[j], ss[i]
	})
	return ss[:n]
}

func chooseL(n int, locs []Location) []Location {
	ss := append([]Location{}, locs...)
	rand.Shuffle(len(ss), func(i, j int) {
		ss[i], ss[j] = ss[j], ss[i]
	})
	return ss[:n]
}
