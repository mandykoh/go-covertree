package covertree

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

var pointDistanceCalls = 0

type Point [3]float64

func (p Point) Distance(other Item) float64 {
	pointDistanceCalls++

	op := other.(*Point)

	total := 0.0
	for i := 0; i < len(op); i++ {
		diff := op[i] - p[i]
		total += diff * diff
	}

	return math.Sqrt(total)
}

func TestTree(t *testing.T) {

	t.Run("Insert()", func(t *testing.T) {

		t.Run("returns the original item when inserting a duplicate", func(t *testing.T) {
			tree := NewInMemoryTree()
			store := tree.store.(*InMemoryStore)

			p1 := randomPoint()
			inserted, err := tree.Insert(&p1)
			if err != nil {
				t.Fatalf("Error inserting point into tree: %v", err)
			}

			p2 := p1
			inserted, err = tree.Insert(&p2)
			if err != nil {
				t.Fatalf("Error inserting point into tree: %v", err)
			}

			nodeCount := traverseTree(tree, store, false)

			if expected, actual := 1, nodeCount; expected != actual {
				t.Errorf("Expected only one node in tree after inserting duplicate but found %d", actual)
			}
			if expected, actual := &p1, inserted; expected != actual {
				t.Errorf("Expected duplicate insertion to return original point but got a different point instead")
			}
		})
	})

	t.Run("with randomly populated tree", func(t *testing.T) {
		tree := NewInMemoryTree()
		store := tree.store.(*InMemoryStore)

		seed := time.Now().UnixNano()
		fmt.Println("Seed:", seed)
		rand.Seed(seed)

		points := randomPoints(10000)

		fmt.Printf("Inserting %d points\n", len(points))
		err := insertPoints(points, tree)
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

				query := randomPoint()
				fmt.Printf("Query point %v (maxResults: %d, maxDistance: %g)\n", query, maxResults, maxDistance)

				resetPointDistanceCalls()

				startTime := time.Now()
				coverTreeResults, err := tree.FindNearest(&query, maxResults, maxDistance)
				finishTime := time.Now()

				if err != nil {
					t.Fatalf("Error querying tree: %v", err)
				}

				coverTreeDistanceCalls := getPointDistanceCalls()

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

		t.Run("FindNearest()", func(t *testing.T) {
			t.Run("returns correct results for nearest neighbour query", func(t *testing.T) {
				compareWithLinearSearch(10, 1, math.MaxFloat64)
			})

			t.Run("returns correct results for k-nearest neighbour query", func(t *testing.T) {
				compareWithLinearSearch(10, 8, math.MaxFloat64)
			})

			t.Run("returns correct results for bounded distance query", func(t *testing.T) {
				compareWithLinearSearch(10, 1, 25)
			})

			t.Run("returns correct results for k-nearest bounded distance query", func(t *testing.T) {
				compareWithLinearSearch(10, 8, 50)
			})
		})
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

func getPointDistanceCalls() int {
	return pointDistanceCalls
}

func linearSearch(query *Point, points []Point, maxResults int, maxDistance float64) (results []ItemWithDistance, distanceCallCount int) {
	resetPointDistanceCalls()

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

	linearSearchDistanceCalls := getPointDistanceCalls()

	fmt.Printf("Linear FindNearest took %d distance comparisons, %dms\n", linearSearchDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)

	for _, r := range results {
		fmt.Printf("Linear FindNearest: %v (distance %g)\n", *r.Item.(*Point), r.Distance)
	}

	return results, linearSearchDistanceCalls
}

func insertPoints(points []Point, tree *Tree) error {
	resetPointDistanceCalls()

	startTime := time.Now()

	for i := range points {
		_, err := tree.Insert(&points[i])
		if err != nil {
			return err
		}
	}

	finishTime := time.Now()

	fmt.Printf("Building tree took %d distance calls, %dms\n", getPointDistanceCalls(), finishTime.Sub(startTime)/time.Millisecond)
	return nil
}

func randomPoint() (point Point) {
	for i := 0; i < len(point); i++ {
		point[i] = rand.Float64() * 1000
	}
	return point
}

func randomPoints(count int) (points []Point) {
	pointsMap := make(map[Point]bool, count)
	for len(pointsMap) < count {
		val := randomPoint()
		if _, exists := pointsMap[val]; !exists {
			pointsMap[val] = true
			points = append(points, val)
		}
	}

	return
}

func resetPointDistanceCalls() {
	pointDistanceCalls = 0
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
		children, _ := store.LoadChildren(item, l)
		for _, c := range children {
			nodeCount += traverseNodes(c, l, indentLevel+1, store, print)
		}
	}

	return
}

func traverseTree(tree *Tree, store *InMemoryStore, print bool) (nodeCount int) {
	return traverseNodes(tree.root, tree.rootLevel, 0, store, print)
}
