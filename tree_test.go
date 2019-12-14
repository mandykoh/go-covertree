package covertree

import (
	"fmt"
	"math"
	"math/rand"
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
				tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)
				_, _ = insertPoints(randomPoints(pointCount), tree)

				for i := 0; i < b.N; i++ {
					p := randomPoint()

					b.StartTimer()
					_ = tree.Insert(&p)
					b.StopTimer()

					_, _ = tree.Remove(&p)
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
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)

			query := randomPoint()
			results, err := tree.FindNearest(&query, 16, math.MaxFloat64)

			if err != nil {
				t.Fatalf("Expected search to succeed but got error: %v", err)
			}
			if expected, actual := 0, len(results); expected != actual {
				t.Errorf("Expected no results but got %d", actual)
			}
		})

		t.Run("returns the only result for tree with a single root node", func(t *testing.T) {
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)
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
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)

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
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)
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
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)
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
			tree, _ := NewTreeWithStore(store, 2, 10.0, distanceBetweenPoints)

			// First point should become the initial root at the level for the specified rootDistance
			p1 := &Point{1.0, 0.0, 0.0}
			err := tree.Insert(p1)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 1, []interface{}{p1}, tree.rootLevel)

			// Second point should be inserted as a child
			p2 := &Point{2.0, 0.0, 0.0}
			err = tree.Insert(p2)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 2, []interface{}{p1}, tree.rootLevel)

			// Third point is very different and should cause a second root to be added
			p3 := &Point{100.0, 0.0, 0.0}
			err = tree.Insert(p3)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 3, []interface{}{p1, p3}, 8)

			// Fourth point is a new child and should deepen the tree a little
			p4 := &Point{1.1, 0.0, 0.0}
			err = tree.Insert(p4)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 4, []interface{}{p1, p3}, 8)

			// Fifth point is another new child at the same depth
			p5 := &Point{2.1, 0.0, 0.0}
			err = tree.Insert(p5)

			if err != nil {
				t.Fatalf("Expected insert to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 5, []interface{}{p1, p3}, 8)
		})

		t.Run("is thread-safe with concurrent reads", func(t *testing.T) {
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)

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

		t.Run("has no effect when the tree is empty", func(t *testing.T) {
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)

			removed, err := tree.Remove(randomPoint())

			if err != nil {
				t.Errorf("Expected removal to have no effect but got error: %v", err)
			}
			if removed != nil {
				t.Errorf("Expected nothing to have been removed but got %v", removed)
			}
		})

		t.Run("removes item from the tree while preserving its children", func(t *testing.T) {
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)

			points := []Point{
				{1.0, 0.0, 0.0},
				{1.1, 0.0, 0.0},
				{1.11, 0.0, 0.0},
				{1.111, 0.0, 0.0},
			}
			_, _ = insertPoints(points, tree)

			removed, _ := tree.Remove(&points[2])

			if expected, actual := &points[2], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}

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

			removed, _ = tree.Remove(&points[1])

			if expected, actual := &points[1], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}

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

			removed, _ = tree.Remove(&points[0])

			if expected, actual := &points[0], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}

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

			removed, _ = tree.Remove(&points[3])

			if expected, actual := &points[3], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}

			nodeCount = traverseTree(tree, tree.store.(*inMemoryStore), false)
			if expected, actual := len(points)-4, nodeCount; expected != actual {
				t.Errorf("Expected %d nodes remaining after removal but found %d", expected, actual)
			}

			// Removed node should no longer be findable (tree is now empty)
			results, _ = tree.FindNearest(&points[3], 1, 0)
			expectSameResults(t, points[3], results, nil)
		})

		t.Run("correctly re-parents orphans that are no longer covered by the root", func(t *testing.T) {
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)

			points := []Point{
				{0.0, 0.0, 0.0},
				{16.0, 0.0, 0.0},
				{15.0, 0.0, 6.0},
			}
			_, _ = insertPoints(points, tree)

			root, rootLevel, _ := loadRoot(tree)

			if expected, actual := &points[0], root; expected != actual {
				t.Errorf("Expected root node to be %v before removal but was %v", expected, actual)
			}
			if expected, actual := 10, rootLevel; expected != actual {
				t.Errorf("Expected root to be at level %d but was %d", expected, actual)
			}

			removed, _ := tree.Remove(&points[1])

			if expected, actual := &points[1], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}

			root, rootLevel, _ = loadRoot(tree)

			nodeCount := traverseTree(tree, tree.store.(*inMemoryStore), false)
			if expected, actual := len(points)-1, nodeCount; expected != actual {
				t.Fatalf("Expected %d nodes remaining after removal but found %d", expected, actual)
			}

			if expected, actual := &points[0], root; expected != actual {
				t.Errorf("Expected root node to be %v after removal but was %v", expected, actual)
			}
			if expected, actual := 10, rootLevel; expected != actual {
				t.Errorf("Expected root to be at level %d but was %d", expected, actual)
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
			tree, _ := NewTreeWithStore(store, 2, 1000.0, distanceBetweenPoints)

			points := []Point{
				{0.0, 0.0, 0.0},
				{16.0, 0.0, 0.0},
				{15.0, 0.0, 1.0},
				{1.0, 0.0, 0.0},
			}
			_, _ = insertPoints(points, tree)

			store.savedCount = 0
			store.expectSavedTree(t, 0, []interface{}{&points[0]}, 4)

			// Removing parent node should cause its uncovered child to be made a new root
			removed, err := tree.Remove(&points[1])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			if expected, actual := &points[1], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}
			store.expectSavedTree(t, 2, []interface{}{&points[0], &points[2]}, 5)

			// Removing leaf node should not affect roots
			removed, err = tree.Remove(&points[3])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			if expected, actual := &points[3], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}
			store.expectSavedTree(t, 3, []interface{}{&points[0], &points[2]}, 5)

			// Removing root node should cause remaining root to become the only root
			removed, err = tree.Remove(&points[0])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			if expected, actual := &points[0], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}
			store.expectSavedTree(t, 4, []interface{}{&points[2]}, 5)

			// Removing final root node should return tree to empty state
			removed, err = tree.Remove(&points[2])
			if err != nil {
				t.Fatalf("Expected removal to succeed but got error: %v", err)
			}
			if expected, actual := &points[2], removed; expected != actual {
				t.Errorf("Expected %v to have been removed but got %v", expected, actual)
			}
			store.expectSavedTree(t, 5, nil, 5)

			// Re-inserting a node should make it a new root
			err = tree.Insert(&points[1])
			if err != nil {
				t.Fatalf("Expected insertion to succeed but got error: %v", err)
			}
			store.expectSavedTree(t, 6, []interface{}{&points[1]}, math.MaxInt32)
		})

		t.Run("allows all remaining nodes to be findable after removal", func(t *testing.T) {
			//rand.Seed(123)
			tree := NewInMemoryTree(2, 1000.0, distanceBetweenPoints)

			points := randomPoints(10)
			_, _ = insertPoints(points, tree)

			var pointsToRemove []interface{}
			for i := range points {
				pointsToRemove = append(pointsToRemove, &points[i])
			}
			rand.Shuffle(len(pointsToRemove), func(i, j int) {
				pointsToRemove[i], pointsToRemove[j] = pointsToRemove[j], pointsToRemove[i]
			})

			for i, p := range pointsToRemove {
				removed, _ := tree.Remove(p)

				if expected, actual := p, removed; expected != actual {
					t.Errorf("Expected %v to have been removed but got %v", expected, actual)
				}

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
		tree, _ := NewTreeWithStore(store, 2, 1000.0, distanceBetweenPointsWithCounter(&distanceCalls))

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

		t.Run("FindNearest()", func(t *testing.T) {

			t.Run("can find all nodes individually", func(t *testing.T) {
				for i := range points {
					results, _ := tree.FindNearest(&points[i], 1, 0)
					expectSameResults(t, points[i], results, []ItemWithDistance{{&points[i], 0}})
				}
			})

			t.Run("returns correct results for nearest neighbour query", func(t *testing.T) {
				for i := 0; i < 10; i++ {
					fmt.Println()
					compareWithLinearSearch(tree, points, 1, math.MaxFloat64, &distanceCalls, t)
				}
			})

			t.Run("returns correct results for k-nearest neighbour query", func(t *testing.T) {
				for i := 0; i < 10; i++ {
					fmt.Println()
					compareWithLinearSearch(tree, points, 8, math.MaxFloat64, &distanceCalls, t)
				}
			})

			t.Run("returns correct results for bounded distance query", func(t *testing.T) {
				for i := 0; i < 10; i++ {
					fmt.Println()
					compareWithLinearSearch(tree, points, 1, 25, &distanceCalls, t)
				}
			})

			t.Run("returns correct results for k-nearest bounded distance query", func(t *testing.T) {
				for i := 0; i < 10; i++ {
					fmt.Println()
					compareWithLinearSearch(tree, points, 8, 50, &distanceCalls, t)
				}
			})
		})
	})
}
