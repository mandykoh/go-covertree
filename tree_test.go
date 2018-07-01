package covertree

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestTree(t *testing.T) {

	t.Run("FindNearest()", func(t *testing.T) {

		t.Run("returns no results for empty tree", func(t *testing.T) {
			tree := NewInMemoryTree(distanceBetweenPoints)

			query := randomPoint()
			results, err := tree.FindNearest(&query, math.MaxInt32, math.MaxFloat64)

			if err != nil {
				t.Fatalf("Expected search to succeed but got error: %v", err)
			}
			if expected, actual := 0, len(results); expected != actual {
				t.Errorf("Expected no results but got %d", actual)
			}
		})

		t.Run("returns the only result for tree with a single root node", func(t *testing.T) {
			tree := NewInMemoryTree(distanceBetweenPoints)
			p := randomPoint()
			tree.Insert(&p)

			query := randomPoint()
			results, err := tree.FindNearest(&query, 2, math.MaxFloat64)

			if err != nil {
				t.Fatalf("Expected search to succeed but got error: %v", err)
			}
			expectSameResults(t, query, results, []ItemWithDistance{{&p, distanceBetweenPoints(&p, &query)}})
		})

		t.Run("with a populated tree", func(t *testing.T) {
			tree := NewInMemoryTree(distanceBetweenPoints)

			points := []Point{
				{1.0, 0.0, 0.0},
				{2.0, 0.0, 0.0},
				{3.0, 0.0, 0.0},
			}
			insertPoints(points, tree)

			t.Run("returns available results when less than the maximum requested", func(t *testing.T) {
				query := Point{0.0, 0.0, 0.0}
				results, err := tree.FindNearest(&query, 5, math.MaxFloat64)

				if err != nil {
					t.Fatalf("Expected search to succeed but got error: %v", err)
				}
				expectSameResults(t, query, results, []ItemWithDistance{
					{&points[0], distanceBetweenPoints(&points[0], &query)},
					{&points[1], distanceBetweenPoints(&points[1], &query)},
					{&points[2], distanceBetweenPoints(&points[2], &query)},
				})
			})

			t.Run("returns up to the maximum requested results", func(t *testing.T) {
				query := Point{0.0, 0.0, 0.0}
				results, err := tree.FindNearest(&query, 2, math.MaxFloat64)

				if err != nil {
					t.Fatalf("Expected search to succeed but got error: %v", err)
				}
				expectSameResults(t, query, results, []ItemWithDistance{
					{&points[0], distanceBetweenPoints(&points[0], &query)},
					{&points[1], distanceBetweenPoints(&points[1], &query)},
				})
			})

			t.Run("returns results up to the maximum requested distance", func(t *testing.T) {
				query := Point{0.0, 0.0, 0.0}
				results, err := tree.FindNearest(&query, 3, 2.0)

				if err != nil {
					t.Fatalf("Expected search to succeed but got error: %v", err)
				}
				expectSameResults(t, query, results, []ItemWithDistance{
					{&points[0], distanceBetweenPoints(&points[0], &query)},
					{&points[1], distanceBetweenPoints(&points[1], &query)},
				})
			})
		})
	})

	t.Run("Insert()", func(t *testing.T) {

		t.Run("returns the original item when inserting a duplicate", func(t *testing.T) {
			tree := NewInMemoryTree(distanceBetweenPoints)
			store := tree.store.(*inMemoryStore)

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

		t.Run("saves the tree meta state when it changes", func(t *testing.T) {
			store := newTestStore()
			tree, _ := NewTreeFromStore(store, distanceBetweenPoints)

			// First point should become the initial root at infinity
			p1 := &Point{1.0, 0.0, 0.0}
			_, err := tree.Insert(p1)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 1, p1, math.MaxInt32)

			// Second point should be inserted as a child, establishing the initial levels
			p2 := &Point{2.0, 0.0, 0.0}
			_, err = tree.Insert(p2)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 2, p1, 1)

			// Third point is very different and should cause re-parenting, but the tree depth should remain the same
			p3 := &Point{100.0, 0.0, 0.0}
			_, err = tree.Insert(p3)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, p3, 7)

			// Fourth point is a new child and should deepen the tree a little without affecting metadata
			p4 := &Point{1.1, 0.0, 0.0}
			_, err = tree.Insert(p4)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, p3, 7)

			// Fifth point is another new child at the same depth and also should not cause an update
			p5 := &Point{2.1, 0.0, 0.0}
			_, err = tree.Insert(p5)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, p3, 7)
		})
	})

	t.Run("with randomly populated tree", func(t *testing.T) {
		distanceCalls := 0
		tree := NewInMemoryTree(distanceBetweenPointsWithCounter(&distanceCalls))
		store := tree.store.(*inMemoryStore)

		seed := time.Now().UnixNano()
		fmt.Println("Seed:", seed)
		rand.Seed(seed)

		points := randomPoints(10000)

		fmt.Printf("Inserting %d points\n", len(points))
		timeTaken, err := insertPoints(points, tree)
		if err != nil {
			t.Fatalf("Error inserting point: %v", err)
		}
		fmt.Printf("Building tree took %d distance calls, %dms\n", distanceCalls, timeTaken/time.Millisecond)

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

				distanceCalls = 0

				startTime := time.Now()
				coverTreeResults, err := tree.FindNearest(&query, maxResults, maxDistance)
				finishTime := time.Now()

				if err != nil {
					t.Fatalf("Error querying tree: %v", err)
				}

				coverTreeDistanceCalls := distanceCalls

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

func distanceBetweenPoints(a, b Item) float64 {
	p1 := a.(*Point)
	p2 := b.(*Point)

	total := 0.0
	for i := 0; i < len(p1); i++ {
		diff := p2[i] - p1[i]
		total += diff * diff
	}

	return math.Sqrt(total)
}

func distanceBetweenPointsWithCounter(counter *int) DistanceFunc {
	return func(a, b Item) float64 {
		*counter++
		return distanceBetweenPoints(a, b)
	}
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

func insertPoints(points []Point, tree *Tree) (timeTaken time.Duration, err error) {
	startTime := time.Now()

	for i := range points {
		_, err := tree.Insert(&points[i])
		if err != nil {
			return 0, err
		}
	}

	finishTime := time.Now()

	return finishTime.Sub(startTime), nil
}

func linearSearch(query *Point, points []Point, maxResults int, maxDistance float64) (results []ItemWithDistance, distanceCallCount int) {
	distanceCalls := 0
	distanceBetween := distanceBetweenPointsWithCounter(&distanceCalls)

	startTime := time.Now()

	results = make([]ItemWithDistance, maxResults, maxResults)

	for i := range points {
		dist := distanceBetween(query, &points[i])
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

	linearSearchDistanceCalls := distanceCalls

	fmt.Printf("Linear FindNearest took %d distance comparisons, %dms\n", linearSearchDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)

	for _, r := range results {
		fmt.Printf("Linear FindNearest: %v (distance %g)\n", *r.Item.(*Point), r.Distance)
	}

	return results, linearSearchDistanceCalls
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

func traverseNodes(item Item, level int, indentLevel int, store *inMemoryStore, print bool) (nodeCount int) {
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
		children, _ := store.LoadChildren(item)
		for _, c := range children.itemsForLevel(l) {
			nodeCount += traverseNodes(c, l, indentLevel+1, store, print)
		}
	}

	return
}

func traverseTree(tree *Tree, store *inMemoryStore, print bool) (nodeCount int) {
	return traverseNodes(tree.root, tree.rootLevel, 0, store, print)
}

type testStore struct {
	inMemoryStore
	savedCount     int
	savedRoot      Item
	savedRootLevel int
}

func newTestStore() *testStore {
	return &testStore{inMemoryStore: *newInMemoryStore()}
}

func (ts *testStore) SaveTree(root Item, rootLevel int) error {
	ts.savedCount++
	ts.savedRoot = root
	ts.savedRootLevel = rootLevel
	return nil
}

func (ts *testStore) expectSavedTree(t *testing.T, saveCount int, root Item, rootLevel int) {
	t.Helper()

	if expected, actual := saveCount, ts.savedCount; expected != actual {
		t.Errorf("Expected tree to have been saved %d times but was saved %d times instead", expected, actual)
	}
	if expected, actual := root, ts.savedRoot; expected != actual {
		t.Errorf("Expected tree root to be %v but was %v", expected, actual)
	}
	if expected, actual := rootLevel, ts.savedRootLevel; expected != actual {
		t.Errorf("Expected tree root level to be at %d but was %d", expected, actual)
	}
}

type Point [3]float64
