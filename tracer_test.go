package covertree

import (
	"math"
	"testing"
	"time"
)

type slowInMemoryStore struct {
	realStore *inMemoryStore
}

func (s *slowInMemoryStore) AddItem(item, parent interface{}, level int) error {
	time.Sleep(1 * time.Millisecond)
	return s.realStore.AddItem(item, parent, level)
}

func (s *slowInMemoryStore) LoadChildren(parents ...interface{}) ([]LevelsWithItems, error) {
	time.Sleep(1 * time.Millisecond)
	return s.realStore.LoadChildren(parents...)
}

func (s *slowInMemoryStore) RemoveItem(item, parent interface{}, level int) error {
	time.Sleep(1 * time.Millisecond)
	return s.realStore.RemoveItem(item, parent, level)
}

func (s *slowInMemoryStore) UpdateItem(item, parent interface{}, level int) error {
	time.Sleep(1 * time.Millisecond)
	return s.realStore.UpdateItem(item, parent, level)
}

func newSlowInMemoryStore() *slowInMemoryStore {
	return &slowInMemoryStore{
		realStore: NewInMemoryStore(distanceBetweenPoints),
	}
}

func TestTracer(t *testing.T) {

	t.Run("FindNearest()", func(t *testing.T) {
		store := newSlowInMemoryStore()
		tree, _ := NewTreeWithStore(store, 2, 5.0, distanceBetweenPoints)

		points := []Point{
			{1.0, 0.0, 0.0},
			{2.0, 0.0, 0.0},
			{4.0, 0.0, 0.0},
		}

		_, err := insertPoints(points, tree)
		if err != nil {
			t.Fatalf("Error inserting point: %v", err)
		}

		traverseTree(tree, store.realStore, false)

		tracer := tree.NewTracer()

		t.Run("records the total covered set size", func(t *testing.T) {
			_, err := tracer.FindNearest(&Point{2.0, 0.0, 0.0}, 1, 0.0)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}

			_, err = tracer.FindNearest(&Point{3.0, 0.0, 0.0}, 2, math.MaxFloat64)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the maximum cover set size", func(t *testing.T) {
			_, err := tracer.FindNearest(&Point{2.0, 0.0, 0.0}, 1, 0.0)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 2, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}

			_, err = tracer.FindNearest(&Point{3.0, 0.0, 0.0}, 2, math.MaxFloat64)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the maximum traversal depth", func(t *testing.T) {
			_, err := tracer.FindNearest(&Point{4.0, 0.0, 0.0}, 1, 0.0)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 4, tracer.MaxLevelsTraversed; expected != actual {
				t.Errorf("Expected maximum traversal depth recorded to be %d but was %d", expected, actual)
			}

			_, err = tracer.FindNearest(&Point{3.0, 0.0, 0.0}, 2, math.MaxFloat64)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 5, tracer.MaxLevelsTraversed; expected != actual {
				t.Errorf("Expected maximum traversal depth recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the store LoadChildren statistics", func(t *testing.T) {
			_, err := tracer.FindNearest(&Point{4.0, 0.0, 0.0}, 1, 0.0)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.LoadChildrenCount; expected != actual {
				t.Errorf("Expected count of LoadChildren operations to be %d but was %d", expected, actual)
			}
			if tracer.TotalLoadChildrenTime == 0 {
				t.Errorf("Expected total time of LoadChildren operations to be non-zero but was zero")
			}

			_, err = tracer.FindNearest(&Point{3.0, 0.0, 0.0}, 2, math.MaxFloat64)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 4, tracer.LoadChildrenCount; expected != actual {
				t.Errorf("Expected count of LoadChildren operations to be %d but was %d", expected, actual)
			}
			if tracer.TotalLoadChildrenTime == 0 {
				t.Errorf("Expected total time of LoadChildren operations to be non-zero but was zero")
			}
		})

		t.Run("records the total search time", func(t *testing.T) {
			_, err := tracer.FindNearest(&Point{4.0, 0.0, 0.0}, 2, math.MaxFloat64)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if tracer.TotalTime == 0 {
				t.Errorf("Expected total time recorded to be non-zero but was zero")
			}
		})
	})

	t.Run("Insert()", func(t *testing.T) {
		var store *slowInMemoryStore
		var tree *Tree

		setup := func() *Tracer {
			store = newSlowInMemoryStore()
			tree, _ = NewTreeWithStore(store, 2, 5.0, distanceBetweenPoints)
			return tree.NewTracer()
		}

		t.Run("records the total covered set size", func(t *testing.T) {
			tracer := setup()

			err := tracer.Insert(&Point{3.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 0, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}

			err = tracer.Insert(&Point{4.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 1, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}

			err = tracer.Insert(&Point{5.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 1, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}

			err = tracer.Insert(&Point{4.4, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 2, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the maximum cover set size", func(t *testing.T) {
			tracer := setup()

			err := tracer.Insert(&Point{3.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 0, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}

			err = tracer.Insert(&Point{4.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 1, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}

			err = tracer.Insert(&Point{5.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 1, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}

			err = tracer.Insert(&Point{4.4, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 2, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the maximum traversal depth", func(t *testing.T) {
			tracer := setup()

			err := tracer.Insert(&Point{2.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 1, tracer.MaxLevelsTraversed; expected != actual {
				t.Errorf("Expected maximum traversal depth recorded to be %d but was %d", expected, actual)
			}

			err = tracer.Insert(&Point{4.41, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.MaxLevelsTraversed; expected != actual {
				t.Errorf("Expected maximum traversal depth recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the store LoadChildren statistics", func(t *testing.T) {
			tracer := setup()

			err := tracer.Insert(&Point{2.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 1, tracer.LoadChildrenCount; expected != actual {
				t.Errorf("Expected count of LoadChildren operations to be %d but was %d", expected, actual)
			}
			if tracer.TotalLoadChildrenTime == 0 {
				t.Errorf("Expected total time of LoadChildren operations to be non-zero but was zero")
			}

			err = tracer.Insert(&Point{4.41, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 2, tracer.LoadChildrenCount; expected != actual {
				t.Errorf("Expected count of LoadChildren operations to be %d but was %d", expected, actual)
			}
			if tracer.TotalLoadChildrenTime == 0 {
				t.Errorf("Expected total time of LoadChildren operations to be non-zero but was zero")
			}
		})

		t.Run("records the total insertion time", func(t *testing.T) {
			tracer := setup()

			err := tracer.Insert(&Point{3.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			err = tracer.Insert(&Point{4.42, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if tracer.TotalTime == 0 {
				t.Errorf("Expected total time recorded to be non-zero but was zero")
			}
		})
	})

	t.Run("Remove()", func(t *testing.T) {
		var store *slowInMemoryStore
		var tree *Tree

		setup := func(t *testing.T) *Tracer {
			t.Helper()

			store = newSlowInMemoryStore()
			tree, _ = NewTreeWithStore(store, 2, 32.0, distanceBetweenPoints)

			points := []Point{
				{1.0, 0.0, 0.0},
				{2.0, 0.0, 0.0},
				{4.0, 0.0, 0.0},
				{8.0, 0.0, 0.0},
				{16.0, 0.0, 0.0},
			}
			_, err := insertPoints(points, tree)
			if err != nil {
				t.Fatalf("Error inserting point: %v", err)
			}

			return tree.NewTracer()
		}

		t.Run("records the total covered set size", func(t *testing.T) {
			tracer := setup(t)

			_, err := tracer.Remove(&Point{2.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 5, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}

			_, err = tracer.Remove(&Point{8.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.TotalCoveredSetSize; expected != actual {
				t.Errorf("Expected total covered set size recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the maximum cover set size", func(t *testing.T) {
			tracer := setup(t)

			_, err := tracer.Remove(&Point{2.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 2, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}

			_, err = tracer.Remove(&Point{8.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.MaxCoverSetSize; expected != actual {
				t.Errorf("Expected maximum cover set size recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the maximum traversal depth", func(t *testing.T) {
			tracer := setup(t)

			_, err := tracer.Remove(&Point{4.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 5, tracer.MaxLevelsTraversed; expected != actual {
				t.Errorf("Expected maximum traversal depth recorded to be %d but was %d", expected, actual)
			}

			_, err = tracer.Remove(&Point{16.0000001, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 4, tracer.MaxLevelsTraversed; expected != actual {
				t.Errorf("Expected maximum traversal depth recorded to be %d but was %d", expected, actual)
			}
		})

		t.Run("records the store LoadChildren statistics", func(t *testing.T) {
			tracer := setup(t)

			_, err := tracer.Remove(&Point{4.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 5, tracer.LoadChildrenCount; expected != actual {
				t.Errorf("Expected count of LoadChildren operations to be %d but was %d", expected, actual)
			}
			if tracer.TotalLoadChildrenTime == 0 {
				t.Errorf("Expected total time of LoadChildren operations to be non-zero but was zero")
			}

			_, err = tracer.Remove(&Point{16.0000001, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := 3, tracer.LoadChildrenCount; expected != actual {
				t.Errorf("Expected count of LoadChildren operations to be %d but was %d", expected, actual)
			}
			if tracer.TotalLoadChildrenTime == 0 {
				t.Errorf("Expected total time of LoadChildren operations to be non-zero but was zero")
			}
		})

		t.Run("records the total removal time", func(t *testing.T) {
			tracer := setup(t)

			_, err := tracer.Remove(&Point{16.0, 0.0, 0.0})
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if tracer.TotalTime == 0 {
				t.Errorf("Expected total time recorded to be non-zero but was zero")
			}
		})
	})
}
