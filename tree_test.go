package covertree

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

var distanceCalls = 0

type Point struct {
	x float64
}

func (p Point) CoverTreeID() string {
	return fmt.Sprintf("%g", p.x)
}

func (p Point) Distance(other Item) float64 {
	distanceCalls++
	op := other.(*Point)
	return math.Abs(op.x - p.x)
}

func PrintTree(item Item, level int, indentLevel int, store *InMemoryStore) (count int) {
	fmt.Printf("%4d: ", level)
	for i := 0; i < indentLevel; i++ {
		fmt.Print("..")
	}
	if indentLevel > 0 {
		fmt.Print(" ")
	}

	fmt.Println(item.CoverTreeID())
	count = 1

	var levels []int
	for k := range store.levelsFor(item) {
		levels = append(levels, k)
	}
	sort.Ints(levels)

	for i := len(levels) - 1; i >= 0; i-- {
		l := levels[i]
		children, _ := store.Load(item, l)
		for _, c := range children {
			count += PrintTree(c, l, indentLevel+1, store)
		}
	}

	return
}

func TestSomething(t *testing.T) {

	store := &InMemoryStore{}

	root := &Point{10}

	tree := &Tree{}
	tree.Insert(root, store)

	for i := 1; i < 20; i++ {
		val := float64(i)/10.0 + 1
		err := tree.Insert(&Point{val}, store)
		fmt.Println("Result", err)
	}

	fmt.Println(tree.Insert(&Point{1000}, store))

	PrintTree(tree.root, 10, 0, store)

	distanceCalls = 0

	query := &Point{6.3}
	results, _ := tree.FindNearest(query, store)

	fmt.Printf("FindNearest took %d distance calls\n", distanceCalls)

	for _, r := range results {
		fmt.Printf("Nearest: %v (distance %.1f)\n", r, r.Distance(query))
	}
}

func TestRandom(t *testing.T) {
	store := &InMemoryStore{}
	tree := &Tree{}

	seed := time.Now().UnixNano()
	fmt.Println("Seed:", seed)
	rand.Seed(seed)

	values := make(map[Point]bool)
	for i := 0; i < 10000000; i++ {
		val := Point{rand.Float64() * 1000}
		values[val] = true
	}

	fmt.Printf("Inserting %d values\n", len(values))

	for k := range values {
		val := k
		err := tree.Insert(&val, store)
		if err != nil {
			fmt.Printf("Error inserting %v: %v\n", k, err)
		}
	}

	//nodeCount := PrintTree(tree.root, tree.rootLevel, 0, store)
	//fmt.Printf("Found %d nodes in tree\n", nodeCount)

	query := &Point{rand.Float64() * 1000}
	fmt.Printf("Query point %v\n", *query)

	distanceCalls = 0
	startTime := time.Now()

	results, _ := tree.FindNearest(query, store)

	finishTime := time.Now()

	fmt.Printf("FindNearest took %d distance comparisons, %dms\n", distanceCalls, finishTime.Sub(startTime)/time.Millisecond)

	for _, r := range results {
		point := *(r.(*Point))
		fmt.Printf("FindNearest: %v (distance %g)\n", point, r.Distance(query))
	}

	distanceCalls = 0
	startTime = time.Now()

	var nearest Point
	var nearestDist float64
	nearestSet := false

	for k := range values {
		val := k
		if !nearestSet {
			nearestSet = true
			nearest = k
			nearestDist = query.Distance(&val)

		} else {
			dist := query.Distance(&val)
			if dist < nearestDist {
				nearest = k
				nearestDist = dist
			}
		}
	}

	finishTime = time.Now()

	fmt.Printf("Linear Nearest took %d distance comparisons, %dms\n", distanceCalls, finishTime.Sub(startTime)/time.Millisecond)
	fmt.Printf("Linear Nearest: %v (distance %g)\n", nearest, nearestDist)
}
