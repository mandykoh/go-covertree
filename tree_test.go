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

	points := randomPoints(10000)

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

	compareWithLinearSearch := func(iterations int, maxResults int, maxDistance float64) {
		t.Helper()

		for n := 0; n < iterations; n++ {
			fmt.Println()

			query := randomPoint(1000)
			fmt.Printf("Query point %v\n", query)

			resetDistanceCalls()

			startTime := time.Now()
			coverTreeResults, err := tree.FindNearest(&query, store, maxResults, maxDistance)
			finishTime := time.Now()

			if err != nil {
				t.Fatalf("Error querying tree: %v", err)
			}

			coverTreeDistanceCalls := getDistanceCalls()

			fmt.Printf("Cover Tree FindNearest took %d distance comparisons, %dms\n", coverTreeDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)
			for _, r := range coverTreeResults {
				point := *(r.Item.(*Point))
				fmt.Printf("Cover Tree FindNearest: %v (distance %g)\n", point, r.Distance)
			}

			linearSearchResults, linearSearchDistanceCalls := linearSearch(&query, points, maxResults, maxDistance)

			expectSameResults(t, query, coverTreeResults, linearSearchResults)

			if coverTreeDistanceCalls >= linearSearchDistanceCalls {
				t.Errorf("Expected cover tree search to require fewer than %d distance comparisons (linear search) but got %d", linearSearchDistanceCalls, coverTreeDistanceCalls)
			}
		}
	}

	t.Run("with nearest neighbour query", func(t *testing.T) {
		compareWithLinearSearch(5, 1, math.MaxFloat64)
	})

	t.Run("with k-nearest neighbour query", func(t *testing.T) {
		compareWithLinearSearch(5, 8, math.MaxFloat64)
	})

	t.Run("with bounded distance query", func(t *testing.T) {
		compareWithLinearSearch(5, 1, 25)
	})

	t.Run("with k-nearest bounded distance query", func(t *testing.T) {
		compareWithLinearSearch(5, 8, 50)
	})
}

func expectSameResults(t *testing.T, query Point, actualResults []ItemWithDistance, expectedResults []ItemWithDistance) {
	t.Helper()

	if expected, actual := len(expectedResults), len(actualResults); expected != actual {
		t.Errorf("Expected %d results but got %d instead", expected, actual)

	}

	availableResults := len(expectedResults)
	if len(actualResults) < availableResults {
		availableResults = len(actualResults)
	}

	for i := 0; i < availableResults; i++ {
		expectedResult := expectedResults[i]
		actualResult := actualResults[i]

		if expected, actual := expectedResult.Item, actualResult.Item; expected != actual {
			t.Errorf("Expected nearest point %d to %v to be %v but got %v", i, query, *expected.(*Point), *actual.(*Point))
		}
		if expected, actual := expectedResult.Distance, actualResult.Distance; expected != actual {
			t.Errorf("Expected distance of nearest point %d to %v to be %v but got %v", i, query, expected, actual)
		}
	}
}

func getDistanceCalls() int32 {
	return atomic.LoadInt32(&distanceCalls)
}

func linearSearch(query *Point, points []Point, maxResults int, maxDistance float64) (results []ItemWithDistance, distanceCallCount int32) {
	resetDistanceCalls()

	startTime := time.Now()

	results = make([]ItemWithDistance, maxResults, maxResults)

	for i := range points {
		dist := query.Distance(&points[i])
		if dist > maxDistance {
			continue
		}

		for j := 0; j < len(results); j++ {
			if results[j].Item == nil || dist < results[j].Distance {
				for k := len(results) - 1; k > j; k-- {
					results[k] = results[k-1]
				}
				results[j].Item = &points[i]
				results[j].Distance = dist
				break
			}
		}
	}

	lastNonNil := len(results) - 1
	for lastNonNil >= 0 && results[lastNonNil].Item == nil {
		lastNonNil--
	}
	results = results[:lastNonNil+1]

	finishTime := time.Now()

	linearSearchDistanceCalls := getDistanceCalls()

	fmt.Printf("Linear FindNearest took %d distance comparisons, %dms\n", linearSearchDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)

	for _, r := range results {
		fmt.Printf("Linear FindNearest: %v (distance %g)\n", *r.Item.(*Point), r.Distance)
	}

	return results, linearSearchDistanceCalls
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
