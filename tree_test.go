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

func BenchmarkTree(b *testing.B) {

	b.Run("Insert()", func(b *testing.B) {
		rand.Seed(123)

		cases := []int{
			100,
			1000,
			10000,
			100000,
		}

		for _, pointCount := range cases {

			b.Run(fmt.Sprintf("with tree of size %d", pointCount), func(b *testing.B) {
				b.StopTimer()
				tree := NewInMemoryTree(2, distanceBetweenPoints)
				_, _ = insertPoints(randomPoints(pointCount), tree)

				for i := 0; i < b.N; i++ {
					p := randomPoint()

					b.StartTimer()
					_ = tree.Insert(&p)
					b.StopTimer()

					_ = tree.Remove(&p)
				}
			})
		}
	})
}

func TestTree(t *testing.T) {

	seed := time.Now().UnixNano()
	fmt.Println("Seed:", seed)
	rand.Seed(seed)

	t.Run("FindNearest()", func(t *testing.T) {

		t.Run("returns no results for empty tree", func(t *testing.T) {
			tree := NewInMemoryTree(2, distanceBetweenPoints)

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
			tree := NewInMemoryTree(2, distanceBetweenPoints)
			p := randomPoint()
			_ = tree.Insert(&p)

			query := randomPoint()
			results, err := tree.FindNearest(&query, 2, math.MaxFloat64)

			if err != nil {
				t.Fatalf("Expected search to succeed but got error: %v", err)
			}
			expectSameResults(t, query, results, []ItemWithDistance{{&p, distanceBetweenPoints(&p, &query)}})
		})

		t.Run("with a populated tree", func(t *testing.T) {
			tree := NewInMemoryTree(2, distanceBetweenPoints)

			points := []Point{
				{1.0, 0.0, 0.0},
				{2.0, 0.0, 0.0},
				{3.0, 0.0, 0.0},
			}
			_, _ = insertPoints(points, tree)

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

		t.Run("inserts duplicates of the root as children of the original item at the root", func(t *testing.T) {
			tree := NewInMemoryTree(2, distanceBetweenPoints)
			store := tree.store.(*inMemoryStore)

			p1 := randomPoint()
			err := tree.Insert(&p1)
			if err != nil {
				t.Fatalf("Error inserting point into tree: %v", err)
			}

			p2 := p1
			err = tree.Insert(&p2)
			if err != nil {
				t.Fatalf("Error inserting point into tree: %v", err)
			}

			nodeCount := traverseTree(tree, store, false)

			if expected, actual := 2, nodeCount; expected != actual {
				t.Errorf("Expected %d nodes in tree after inserting duplicate but found %d", expected, actual)
			}

			found, err := tree.FindNearest(&p2, 2, 0.0)
			if err != nil {
				t.Fatalf("Expected lookup of duplicates to succeed but got error: %v", err)
			}
			if expected, actual := 2, len(found); expected != actual {
				t.Errorf("Expected %d duplicate items to be findable but found %d instead", expected, actual)
			} else {
				if expected, actual := &p1, found[0].Item; expected != actual {
					t.Errorf("Expected first inserted duplicate to be findable but got %v", actual)
				}
				if expected, actual := &p2, found[1].Item; expected != actual {
					t.Errorf("Expected second inserted duplicate to be findable but got %v", actual)
				}
			}
		})

		t.Run("inserts duplicates as children of the original item", func(t *testing.T) {
			tree := NewInMemoryTree(2, distanceBetweenPoints)
			store := tree.store.(*inMemoryStore)

			_, _ = insertPoints(randomPoints(2), tree)

			p1 := randomPoint()
			err := tree.Insert(&p1)
			if err != nil {
				t.Fatalf("Error inserting point into tree: %v", err)
			}

			p2 := p1
			err = tree.Insert(&p2)
			if err != nil {
				t.Fatalf("Error inserting point into tree: %v", err)
			}

			nodeCount := traverseTree(tree, store, true)

			if expected, actual := 4, nodeCount; expected != actual {
				t.Errorf("Expected %d nodes in tree after inserting duplicate but found %d", expected, actual)
			}

			found, err := tree.FindNearest(&p2, 2, 0.0)
			if err != nil {
				t.Fatalf("Expected lookup of duplicates to succeed but got error: %v", err)
			}
			if expected, actual := 2, len(found); expected != actual {
				t.Errorf("Expected %d duplicate items to be findable but found %d instead", expected, actual)
			} else {
				if expected, actual := &p1, found[0].Item; expected != actual {
					t.Errorf("Expected first inserted duplicate to be findable but got %v", actual)
				}
				if expected, actual := &p2, found[1].Item; expected != actual {
					t.Errorf("Expected second inserted duplicate to be findable but got %v", actual)
				}
			}
		})

		t.Run("saves the tree root state when it changes", func(t *testing.T) {
			store := newTestStore(distanceBetweenPoints)
			tree, _ := NewTreeWithStore(store, 2, distanceBetweenPoints)

			// First point should become the initial root at infinity
			p1 := &Point{1.0, 0.0, 0.0}
			err := tree.Insert(p1)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 1, p1, math.MaxInt32)

			// Second point should be inserted as a child, establishing the initial levels
			p2 := &Point{2.0, 0.0, 0.0}
			err = tree.Insert(p2)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 2, p1, 0)

			// Third point is very different and should cause the root to be promoted
			p3 := &Point{100.0, 0.0, 0.0}
			err = tree.Insert(p3)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, p1, 8)

			// Fourth point is a new child and should deepen the tree a little without affecting the root
			p4 := &Point{1.1, 0.0, 0.0}
			err = tree.Insert(p4)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, p1, 8)

			// Fifth point is another new child at the same depth and also should not cause an update
			p5 := &Point{2.1, 0.0, 0.0}
			err = tree.Insert(p5)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, p1, 8)
		})

		t.Run("is thread-safe with concurrent reads", func(t *testing.T) {
			store := newTestStore(distanceBetweenPoints)
			tree, _ := NewTreeWithStore(store, 2, distanceBetweenPoints)

			points := randomPoints(10000)

			const workers = 8
			var insertQueue = make(chan *Point, workers*2)
			var doneGroup sync.WaitGroup
			doneGroup.Add(workers)

			for i := 0; i < workers; i++ {
				go func() {
					for p := range insertQueue {
						_ = tree.Insert(p)
						_, _ = tree.FindNearest(p, 1, 0.0)
					}
					doneGroup.Done()
				}()
			}

			for i := range points {
				insertQueue <- &points[i]
			}
			close(insertQueue)
			doneGroup.Wait()

			for i := range points {
				p := &points[i]

				results, err := tree.FindNearest(p, 1, 0.0)

				if err != nil {
					t.Fatalf("Expected success looking up point but got error: %v", err)
				}

				if len(results) == 0 {
					t.Errorf("Expected point %v to be findable but wasn’t", p)
				} else if results[0].Item != p {
					t.Errorf("Expected point %v to be findable but found %v instead", p, results[0].Item)
				}
			}
		})
	})

	t.Run("Remove()", func(t *testing.T) {

		t.Run("removes item from the tree while preserving its children", func(t *testing.T) {
			tree := NewInMemoryTree(2, distanceBetweenPoints)

			points := []Point{
				{1.0, 0.0, 0.0},
				{1.1, 0.0, 0.0},
				{1.11, 0.0, 0.0},
				{1.111, 0.0, 0.0},
			}
			_, _ = insertPoints(points, tree)

			_ = tree.Remove(&points[2])

			nodeCount := traverseTree(tree, tree.store.(*inMemoryStore), false)
			if expected, actual := len(points)-1, nodeCount; expected != actual {
				t.Errorf("Expected %d nodes remaining after removal but found %d", expected, actual)
			}

			// Removed node should no longer be findable
			results, _ := tree.FindNearest(&points[2], 1, 0)
			expectSameResults(t, points[2], results, nil)

			// Orphaned child node should have been re-parented and still be findable
			results, _ = tree.FindNearest(&points[3], 1, 0)
			expectSameResults(t, points[3], results, []ItemWithDistance{{&points[3], 0}})

			_ = tree.Remove(&points[1])

			nodeCount = traverseTree(tree, tree.store.(*inMemoryStore), false)
			if expected, actual := len(points)-2, nodeCount; expected != actual {
				t.Errorf("Expected %d nodes remaining after removal but found %d", expected, actual)
			}

			// Removed node should no longer be findable
			results, _ = tree.FindNearest(&points[1], 1, 0)
			expectSameResults(t, points[1], results, nil)

			// Orphaned child node should have been re-parented and still be findable
			results, _ = tree.FindNearest(&points[3], 1, 0)
			expectSameResults(t, points[3], results, []ItemWithDistance{{&points[3], 0}})

			_ = tree.Remove(&points[0])

			nodeCount = traverseTree(tree, tree.store.(*inMemoryStore), false)
			if expected, actual := len(points)-3, nodeCount; expected != actual {
				t.Errorf("Expected %d nodes remaining after removal but found %d", expected, actual)
			}

			// Removed node should no longer be findable
			results, _ = tree.FindNearest(&points[0], 1, 0)
			expectSameResults(t, points[0], results, nil)

			// Orphaned child node should have been re-parented and still be findable
			results, _ = tree.FindNearest(&points[3], 1, 0)
			expectSameResults(t, points[3], results, []ItemWithDistance{{&points[3], 0}})

			_ = tree.Remove(&points[3])

			nodeCount = traverseTree(tree, tree.store.(*inMemoryStore), false)
			if expected, actual := len(points)-4, nodeCount; expected != actual {
				t.Errorf("Expected %d nodes remaining after removal but found %d", expected, actual)
			}

			// Removed node should no longer be findable (tree is now empty)
			results, _ = tree.FindNearest(&points[3], 1, 0)
			expectSameResults(t, points[3], results, nil)
		})

		t.Run("correctly re-parents orphans that are no longer covered by the root", func(t *testing.T) {
			tree := NewInMemoryTree(2, distanceBetweenPoints)

			points := []Point{
				{0.0, 0.0, 0.0},
				{16.0, 0.0, 0.0},
				{15.0, 0.0, 6.0},
			}
			_, _ = insertPoints(points, tree)

			root, rootLevel, _ := tree.loadRoot()

			if expected, actual := &points[0], root; expected != actual {
				t.Errorf("Expected root node to be %v before removal but was %v", expected, actual)
			}
			if expected, actual := 4, rootLevel; expected != actual {
				t.Errorf("Expected root to be at level %d but was %d", expected, actual)
			}

			_ = tree.Remove(&points[1])
			root, rootLevel, _ = tree.loadRoot()

			nodeCount := traverseTree(tree, tree.store.(*inMemoryStore), false)
			if expected, actual := len(points)-1, nodeCount; expected != actual {
				t.Fatalf("Expected %d nodes remaining after removal but found %d", expected, actual)
			}

			if expected, actual := &points[0], root; expected != actual {
				t.Errorf("Expected root node to be %v after removal but was %v", expected, actual)
			}
			if expected, actual := 6, rootLevel; expected != actual {
				t.Errorf("Expected root to have been promoted to level %d but was %d", expected, actual)
			}

			// Removed node should no longer be findable
			results, _ := tree.FindNearest(&points[1], 1, 0)
			expectSameResults(t, points[1], results, nil)

			// Remaining nodes should still be findable
			results, _ = tree.FindNearest(&points[0], 1, 0)
			expectSameResults(t, points[0], results, []ItemWithDistance{{&points[0], 0}})
			results, _ = tree.FindNearest(&points[2], 1, 0)
			expectSameResults(t, points[2], results, []ItemWithDistance{{&points[2], 0}})
		})

		t.Run("saves the tree root state when it changes", func(t *testing.T) {
			store := newTestStore(distanceBetweenPoints)
			tree, _ := NewTreeWithStore(store, 2, distanceBetweenPoints)

			points := []Point{
				{0.0, 0.0, 0.0},
				{16.0, 0.0, 0.0},
				{15.0, 0.0, 6.0},
				{1.0, 0.0, 0.0},
			}
			_, _ = insertPoints(points, tree)

			store.savedCount = 0
			store.expectSavedTree(t, 0, &points[0], 4)

			// Removing parent node should cause its uncovered child to bubble up and the root to be promoted
			err := tree.Remove(&points[1])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 1, &points[0], 6)

			// Removing leaf node should not affect metadata
			err = tree.Remove(&points[3])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 1, &points[0], 6)

			// Removing root node should cause child to become the root
			err = tree.Remove(&points[0])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 2, &points[2], 6)

			// Removing final root node should return tree to empty state
			err = tree.Remove(&points[2])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, nil, 6)

			// Re-inserting a node should make it the new root
			err = tree.Insert(&points[1])
			if err != nil {
				t.Fatalf("Expected insertion to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 4, &points[1], math.MaxInt32)
		})

		t.Run("allows all remaining nodes to be findable after removal", func(t *testing.T) {
			tree := NewInMemoryTree(2, distanceBetweenPoints)

			points := randomPoints(100)
			_, _ = insertPoints(points, tree)

			var pointsToRemove []interface{}
			for i := range points {
				pointsToRemove = append(pointsToRemove, &points[i])
			}
			rand.Shuffle(len(pointsToRemove), func(i, j int) {
				pointsToRemove[i], pointsToRemove[j] = pointsToRemove[j], pointsToRemove[i]
			})

			for i, p := range pointsToRemove {
				_ = tree.Remove(p)

				nodeCount := traverseTree(tree, tree.store.(*inMemoryStore), false)
				if expected, actual := len(points)-i-1, nodeCount; expected != actual {
					t.Fatalf("Expected %d nodes remaining after %d removals but found %d", expected, i+1, actual)
				}

				// Removed node should no longer be findable
				results, _ := tree.FindNearest(p, 1, 0)
				expectSameResults(t, *p.(*Point), results, nil)

				// All other nodes should still be findable
				for j := i + 1; j < len(pointsToRemove); j++ {
					results, _ = tree.FindNearest(pointsToRemove[j], 1, 0)
					expectSameResults(t, *pointsToRemove[j].(*Point), results, []ItemWithDistance{{pointsToRemove[j], 0}})
				}
			}
		})
	})

	t.Run("with randomly populated tree", func(t *testing.T) {
		distanceCalls := 0
		store := newInMemoryStore(distanceBetweenPoints)
		tree, _ := NewTreeWithStore(store, 2, distanceBetweenPointsWithCounter(&distanceCalls))

		points := randomPoints(1000)

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

			t.Run("can find all nodes individually", func(t *testing.T) {
				for i := range points {
					results, _ := tree.FindNearest(&points[i], 1, 0)
					expectSameResults(t, points[i], results, []ItemWithDistance{{&points[i], 0}})
				}
			})

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

func distanceBetweenPoints(a, b interface{}) float64 {
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
	return func(a, b interface{}) float64 {
		*counter++
		return distanceBetweenPoints(a, b)
	}
}

func expectSameResults(t *testing.T, query Point, actualResults []ItemWithDistance, expectedResults []ItemWithDistance) {
	t.Helper()

	if expected, actual := len(expectedResults), len(actualResults); expected != actual {
		t.Fatalf("Expected %d results for %v but got %d instead", expected, query, actual)

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
		err := tree.Insert(&points[i])
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

func traverseNodes(item, parent interface{}, level int, indentLevel int, store *inMemoryStore, print bool) (nodeCount int) {
	if print {
		fmt.Printf("%4d: ", level)
		for i := 0; i < indentLevel; i++ {
			fmt.Print("..")
		}
		if indentLevel > 0 {
			fmt.Print(" ")
		}

		fmt.Printf("%v", item)

		if parent != nil {
			fmt.Printf(" (%g)", distanceBetweenPoints(item, parent))
		}

		fmt.Println()
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
		for _, c := range children.itemsAt(l) {
			nodeCount += traverseNodes(c, item, l, indentLevel+1, store, print)
		}
	}

	return
}

func traverseTree(tree *Tree, store *inMemoryStore, print bool) (nodeCount int) {
	if print {
		fmt.Println("---")
	}

	root, rootLevel, _ := tree.loadRoot()

	if root == nil {
		return 0
	}

	return traverseNodes(root, nil, rootLevel, 0, store, print)
}

type testStore struct {
	inMemoryStore
	savedCount     int
	savedRoot      interface{}
	savedRootLevel int
}

func newTestStore(distanceFunc DistanceFunc) *testStore {
	return &testStore{inMemoryStore: *newInMemoryStore(distanceFunc)}
}

func (ts *testStore) AddItem(item, parent interface{}, level int) error {
	if parent == nil {
		ts.savedCount++
		ts.savedRoot = item
		ts.savedRootLevel = level
	}
	return ts.inMemoryStore.AddItem(item, parent, level)
}

func (ts *testStore) UpdateItem(item, parent interface{}, level int) error {
	if parent == nil {
		ts.savedCount++
		ts.savedRoot = item
		ts.savedRootLevel = level
	}
	return ts.inMemoryStore.UpdateItem(item, parent, level)
}

func (ts *testStore) expectSavedTree(t *testing.T, saveCount int, root interface{}, rootLevel int) {
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

func (p *Point) String() string {
	return fmt.Sprintf("[%g %g %g]", p[0], p[1], p[2])
}
