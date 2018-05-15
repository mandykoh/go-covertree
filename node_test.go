package covertree

import (
	"fmt"
	"math"
	"sort"
	"testing"
)

type Point struct {
	x float64
}

func (p Point) CoverTreeID() string {
	return fmt.Sprintf("%g", p.x)
}

func (p Point) Distance(other Item) float64 {
	op := other.(*Point)
	return math.Abs(op.x - p.x)
}

func PrintTree(item Item, level int, indentLevel int, store *InMemoryStore) {
	fmt.Printf("%4d: ", level)
	for i := 0; i < indentLevel; i++ {
		fmt.Print("  ")
	}

	fmt.Println(item.CoverTreeID())

	var levels []int
	for k := range store.levelsFor(item) {
		levels = append(levels, k)
	}
	sort.Ints(levels)

	for i := len(levels) - 1; i >= 0; i-- {
		l := levels[i]
		children, _ := store.Load(item, l)
		for _, c := range children {
			PrintTree(c, l, indentLevel+1, store)
		}
	}
}

func TestSomething(t *testing.T) {

	store := &InMemoryStore{}

	root := &Point{1}

	for i := 1; i < 20; i++ {
		val := float64(i)/10.0 + 1
		ok, err := Insert(&Point{val}, coverSet{root}, 10, store)
		fmt.Println("Result", ok, err)
	}

	PrintTree(root, 10, 0, store)
}
