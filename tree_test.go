package covertree

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"
)

var distanceCalls = 0

type Point [3]float64

func RandomPoint(scale float64) Point {
	p := Point{}
	for i := 0; i < len(p); i++ {
		p[i] = rand.Float64() * scale
	}
	return p
}

func (p Point) Distance(other Item) float64 {
	distanceCalls++
	op := other.(*Point)

	total := 0.0
	for i := 0; i < len(op); i++ {
		diff := op[i] - p[i]
		total += diff * diff
	}

	return math.Sqrt(total)
}

func PrintTree(item Item, level int, indentLevel int, store *InMemoryStore, actuallyPrint bool) (count int) {
	if actuallyPrint {
		fmt.Printf("%4d: ", level)
		for i := 0; i < indentLevel; i++ {
			fmt.Print("..")
		}
		if indentLevel > 0 {
			fmt.Print(" ")
		}

		fmt.Println(item)
	}

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
			count += PrintTree(c, l, indentLevel+1, store, actuallyPrint)
		}
	}

	return
}

//func TestSomething(t *testing.T) {
//
//	store := NewInMemoryStore()
//
//	root := &Point{10, 10, 10}
//
//	tree := &Tree{}
//	tree.Insert(root, store)
//
//	for i := 1; i < 20; i++ {
//		val := float64(i)/10.0 + 1
//		err := tree.Insert(&Point{val, val, val}, store)
//		fmt.Println("Result", err)
//	}
//
//	fmt.Println(tree.Insert(&Point{1000, 1000, 1000}, store))
//
//	PrintTree(tree.root, 10, 0, store)
//
//	distanceCalls = 0
//
//	query := &Point{6.3, 6.3, 6.3}
//	results, _ := tree.FindNearest(query, store)
//
//	fmt.Printf("FindNearest took %d distance calls\n", distanceCalls)
//
//	for _, r := range results {
//		fmt.Printf("Nearest: %v (distance %.1f)\n", r, r.Distance(query))
//	}
//}

func TestRandom(t *testing.T) {
	store := NewInMemoryStore()
	tree := &Tree{}

	seed := time.Now().UnixNano()
	fmt.Println("Seed:", seed)
	rand.Seed(seed)

	var values []Point
	{
		valuesMap := make(map[Point]bool)
		for i := 0; i < 100000; i++ {
			val := RandomPoint(1000)
			valuesMap[val] = true
		}

		for k := range valuesMap {
			values = append(values, k)
		}
	}

	fmt.Printf("Inserting %d values\n", len(values))

	distanceCalls = 0
	startTime := time.Now()

	pointsToInsert := make(chan *Point)

	const insertThreads = 8

	treeReady := sync.WaitGroup{}
	treeReady.Add(insertThreads)
	for i := 0; i < insertThreads; i++ {
		go func() {
			for {
				p, ok := <-pointsToInsert
				if !ok {
					break
				}

				err := tree.Insert(p, store)
				if err != nil {
					fmt.Printf("Error inserting %v: %v\n", values[i], err)
				}
			}

			treeReady.Done()
		}()
	}

	for i := range values {
		pointsToInsert <- &values[i]
	}
	close(pointsToInsert)
	treeReady.Wait()

	finishTime := time.Now()

	fmt.Printf("Building tree took %d distance calls, %dms\n", distanceCalls, finishTime.Sub(startTime)/time.Millisecond)

	nodeCount := PrintTree(tree.root, tree.rootLevel, 0, store, false)
	fmt.Printf("Found %d nodes in tree\n", nodeCount)

	for n := 0; n < 5; n++ {
		fmt.Println()

		query := RandomPoint(1000)
		fmt.Printf("Query point %v\n", query)

		distanceCalls = 0
		startTime = time.Now()

		results, _ := tree.FindNearest(&query, store)

		finishTime = time.Now()

		fmt.Printf("Cover Tree FindNearest took %d distance comparisons, %dms\n", distanceCalls, finishTime.Sub(startTime)/time.Millisecond)

		for _, r := range results {
			point := *(r.(*Point))
			fmt.Printf("Cover Tree FindNearest: %v (distance %g)\n", point, r.Distance(&query))
		}

		distanceCalls = 0
		startTime = time.Now()

		var nearest *Point
		var nearestDist float64
		for i := range values {
			if nearest == nil {
				nearest = &values[i]
				nearestDist = query.Distance(nearest)

			} else {
				dist := query.Distance(&values[i])
				if dist < nearestDist {
					nearest = &values[i]
					nearestDist = dist
				}
			}
		}

		finishTime = time.Now()

		fmt.Printf("Linear FindNearest took %d distance comparisons, %dms\n", distanceCalls, finishTime.Sub(startTime)/time.Millisecond)
		fmt.Printf("Linear FindNearest: %v (distance %g)\n", *nearest, nearestDist)
	}
}
