package covertree

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var distanceCalls = int32(0)

type Point [3]float64

func (p Point) Distance(other Item) float64 {
	atomic.AddInt32(&distanceCalls, 1)

	op := other.(*Point)

	total := 0.0
	for i := 0; i < len(op); i++ {
		diff := op[i] - p[i]
		total += diff * diff
	}

	return math.Sqrt(total)
}

func TestComparisonToLinearSearch(t *testing.T) {
	store := NewInMemoryStore()
	tree := &Tree{}

	seed := time.Now().UnixNano()
	fmt.Println("Seed:", seed)
	rand.Seed(seed)

	points := randomPoints(100000)

	fmt.Printf("Inserting %d points\n", len(points))
	err := insertPoints(points, tree, store)
	if err != nil {
		t.Fatalf("Error inserting point: %v", err)
	}

	nodeCount := traverseTree(tree, store, false)
	fmt.Printf("Found %d nodes in tree\n", nodeCount)
	if expected := len(points); nodeCount != expected {
		t.Fatalf("Expected %d nodes in tree but found %d", expected, nodeCount)
	}

	for n := 0; n < 10; n++ {
		fmt.Println()

		query := randomPoint(1000)
		fmt.Printf("Query point %v\n", query)

		resetDistanceCalls()

		startTime := time.Now()
		results, err := tree.FindNearest(&query, store, 2, 100)
		finishTime := time.Now()

		if err != nil {
			t.Fatalf("Error querying tree: %v", err)
		}

		if len(results) == 0 {
			t.Fatalf("Expected some results but got none")
		}

		coverTreeNearest := results[0]
		coverTreeDistanceCalls := getDistanceCalls()

		fmt.Printf("Cover Tree FindNearest took %d distance comparisons, %dms\n", coverTreeDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)

		for _, r := range results {
			point := *(r.Item.(*Point))
			fmt.Printf("Cover Tree FindNearest: %v (distance %g)\n", point, r.Distance)
		}

		resetDistanceCalls()

		startTime = time.Now()

		var linearSearchNearest *Point
		var linearSearchNearestDist float64
		for i := range points {
			if linearSearchNearest == nil {
				linearSearchNearest = &points[i]
				linearSearchNearestDist = query.Distance(linearSearchNearest)

			} else {
				dist := query.Distance(&points[i])
				if dist < linearSearchNearestDist {
					linearSearchNearest = &points[i]
					linearSearchNearestDist = dist
				}
			}
		}

		finishTime = time.Now()

		linearSearchDistanceCalls := getDistanceCalls()

		fmt.Printf("Linear FindNearest took %d distance comparisons, %dms\n", linearSearchDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)
		fmt.Printf("Linear FindNearest: %v (distance %g)\n", *linearSearchNearest, linearSearchNearestDist)

		if linearSearchNearest != coverTreeNearest.Item {
			t.Errorf("Expected nearest point to %v to be %v but got %v", query, *linearSearchNearest, *coverTreeNearest.Item.(*Point))
		}
		if linearSearchNearestDist != coverTreeNearest.Distance {
			t.Errorf("Expected distance to nearest point to %v to be %v but got %v", query, linearSearchNearestDist, coverTreeNearest.Distance)
		}
		if coverTreeDistanceCalls >= linearSearchDistanceCalls {
			t.Errorf("Expected cover tree search to require fewer than %d distance comparisons (linear search) but got %d", linearSearchDistanceCalls, coverTreeDistanceCalls)
		}
	}
}

func getDistanceCalls() int32 {
	return atomic.LoadInt32(&distanceCalls)
}

func insertPoints(points []Point, tree *Tree, store Store) (err error) {
	const insertThreads = 8

	resetDistanceCalls()

	pointsToInsert := make(chan *Point)
	insertCount := int32(0)

	errored := int32(0)

	treeReady := sync.WaitGroup{}
	treeReady.Add(insertThreads)

	for i := 0; i < insertThreads; i++ {
		go func() {
			for {
				p, ok := <-pointsToInsert
				if !ok {
					break
				}

				insertErr := tree.Insert(p, store)
				if insertErr != nil {
					if atomic.SwapInt32(&errored, 1) == 0 {
						err = insertErr
					}
					break
				}

				if inserted := atomic.AddInt32(&insertCount, 1); inserted%100000 == 0 {
					fmt.Printf("%d to go\n", len(points)-int(inserted))
				}
			}

			treeReady.Done()
		}()
	}

	startTime := time.Now()

	for i := range points {
		if atomic.LoadInt32(&errored) != 0 {
			break
		}
		pointsToInsert <- &points[i]
	}
	close(pointsToInsert)
	treeReady.Wait()

	finishTime := time.Now()

	fmt.Printf("Building tree took %d distance calls, %dms\n", getDistanceCalls(), finishTime.Sub(startTime)/time.Millisecond)
	return nil
}

func randomPoint(scale float64) (point Point) {
	for i := 0; i < len(point); i++ {
		point[i] = rand.Float64() * scale
	}
	return point
}

func randomPoints(count int) (points []Point) {
	pointsMap := make(map[Point]bool, count)
	for len(pointsMap) < count {
		val := randomPoint(1000)
		pointsMap[val] = true
	}

	for k := range pointsMap {
		points = append(points, k)
	}

	return
}

func resetDistanceCalls() {
	atomic.StoreInt32(&distanceCalls, 0)
}

func traverseNodes(item Item, level int, indentLevel int, store *InMemoryStore, print bool) (nodeCount int) {
	if print {
		fmt.Printf("%4d: ", level)
		for i := 0; i < indentLevel; i++ {
			fmt.Print("..")
		}
		if indentLevel > 0 {
			fmt.Print(" ")
		}

		fmt.Println(item)
	}

	nodeCount = 1

	var levels []int
	for k := range store.levelsFor(item) {
		levels = append(levels, k)
	}
	sort.Ints(levels)

	for i := len(levels) - 1; i >= 0; i-- {
		l := levels[i]
		children, _ := store.Load(item, l)
		for _, c := range children {
			nodeCount += traverseNodes(c, l, indentLevel+1, store, print)
		}
	}

	return
}

func traverseTree(tree *Tree, store *InMemoryStore, print bool) (nodeCount int) {
	return traverseNodes(tree.root, tree.rootLevel, 0, store, print)
}
