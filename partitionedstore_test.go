package covertree

import (
	"fmt"
	"testing"
)

func TestPartitionedStore(t *testing.T) {

	partitioningFunc := func(parentItem interface{}) string {
		return fmt.Sprintf("%v", parentItem)
	}

	addPoints := func(points []Point, s Store, t *testing.T) {
		t.Helper()

		for i := range points {
			var parent interface{}
			if i > 0 {
				parent = &points[i-1]
			}

			err := s.AddItem(&points[i], parent, i)
			if err != nil {
				t.Fatalf("Expected item to be added but got error: %v", err)
			}
		}
	}

	countItems := func(items map[interface{}]map[int][]interface{}) int {
		total := 0

		for _, levels := range items {
			for _, children := range levels {
				total += len(children)
			}
		}

		return total
	}

	t.Run("consistently uses the same stores for items", func(t *testing.T) {
		s1 := NewInMemoryStore(distanceBetweenPoints)
		s2 := NewInMemoryStore(distanceBetweenPoints)
		s := NewPartitionedStore(partitioningFunc, s1, s2)

		points := randomPoints(100)

		stores := make([]Store, len(points))
		for i := range points {
			stores[i], _ = s.storeForParent(&points[i])

			if stores[i] != s1 && stores[i] != s2 {
				t.Fatalf("Expected returned store to be one of the provided stores but was %v", stores[i])
			}
		}

		for i := range points {
			store, _ := s.storeForParent(&points[i])
			if store != stores[i] {
				t.Fatalf("Expected point %d to use store %v but was %v", i, stores[i], store)
			}
		}
	})

	t.Run("distributes AddItem operations across underlying stores", func(t *testing.T) {
		s1 := NewInMemoryStore(distanceBetweenPoints)
		s2 := NewInMemoryStore(distanceBetweenPoints)
		s3 := NewInMemoryStore(distanceBetweenPoints)
		s4 := NewInMemoryStore(distanceBetweenPoints)
		s := NewPartitionedStore(partitioningFunc, s1, s2, s3, s4)

		points := randomPoints(100)
		addPoints(points, s, t)

		s1ItemCount := len(s1.items)
		s2ItemCount := len(s2.items)
		s3ItemCount := len(s3.items)
		s4ItemCount := len(s4.items)

		if s1ItemCount == 0 {
			t.Errorf("Expected some points in the first store but found none")
		}
		if s2ItemCount == 0 {
			t.Errorf("Expected some points in the second store but found none")
		}
		if s3ItemCount == 0 {
			t.Errorf("Expected some points in the third store but found none")
		}
		if s4ItemCount == 0 {
			t.Errorf("Expected some points in the fourth store but found none")
		}
		if expected, actual := len(points), s1ItemCount+s2ItemCount+s3ItemCount+s4ItemCount; expected != actual {
			t.Errorf("Expected points in all stores to total %d but was %d", expected, actual)
		}
	})

	t.Run("distributes LoadChildren operations across underlying stores", func(t *testing.T) {
		s1 := NewInMemoryStore(distanceBetweenPoints)
		s2 := NewInMemoryStore(distanceBetweenPoints)
		s3 := NewInMemoryStore(distanceBetweenPoints)
		s4 := NewInMemoryStore(distanceBetweenPoints)
		s := NewPartitionedStore(partitioningFunc, s1, s2, s3, s4)

		points := randomPoints(100)
		addPoints(points, s, t)

		parents := make([]interface{}, len(points))
		for i := range points[:len(points)-1] {
			parents[i+1] = &points[i]
		}

		children, err := s.LoadChildren(parents...)
		if err != nil {
			t.Fatalf("Expected children to be loaded but got error: %v", err)
		}

		if expected, actual := len(points), len(children); expected != actual {
			t.Errorf("Expected children to be returned in %d levels but got %d", expected, actual)
		}
	})

	t.Run("distributes RemoveItem operations across underlying stores", func(t *testing.T) {
		s1 := NewInMemoryStore(distanceBetweenPoints)
		s2 := NewInMemoryStore(distanceBetweenPoints)
		s3 := NewInMemoryStore(distanceBetweenPoints)
		s4 := NewInMemoryStore(distanceBetweenPoints)
		s := NewPartitionedStore(partitioningFunc, s1, s2, s3, s4)

		points := randomPoints(100)
		addPoints(points, s, t)

		for i := range points {
			var parent interface{}
			if i > 0 {
				parent = &points[i-1]
			}

			err := s.RemoveItem(&points[i], parent, i)
			if err != nil {
				t.Fatalf("Expected point to be removed but got error: %v", err)
			}
		}

		if pointCount := countItems(s1.items); pointCount > 0 {
			t.Errorf("Expected first store to be empty but found %d points", pointCount)
		}
		if pointCount := countItems(s2.items); pointCount > 0 {
			t.Errorf("Expected second store to be empty but found %d points", pointCount)
		}
		if pointCount := countItems(s3.items); pointCount > 0 {
			t.Errorf("Expected third store to be empty but found %d points", pointCount)
		}
		if pointCount := countItems(s4.items); pointCount > 0 {
			t.Errorf("Expected fourth store to be empty but found %d points", pointCount)
		}
	})

	t.Run("distributes UpdateItem operations across underlying stores", func(t *testing.T) {
		s1 := NewInMemoryStore(distanceBetweenPoints)
		s2 := NewInMemoryStore(distanceBetweenPoints)
		s3 := NewInMemoryStore(distanceBetweenPoints)
		s4 := NewInMemoryStore(distanceBetweenPoints)
		s := NewPartitionedStore(partitioningFunc, s1, s2, s3, s4)

		points := randomPoints(100)
		addPoints(points, s, t)

		for i := range points {
			err := s.UpdateItem(&points[i], &points[i], 0)
			if err != nil {
				t.Fatalf("Expected point to be updated but got error: %v", err)
			}
		}

		for i := range points {
			children, err := s.LoadChildren(&points[i])
			if err != nil {
				t.Fatalf("Expected children to be loaded but got error: %v", err)
			}

			if expected, actual := 1, len(children); expected != actual {
				t.Fatalf("Expected one set of children for point %d but found %d", i, actual)
			}
			if actual := len(children[0].items); actual == 0 {
				t.Fatalf("Expected at least one level of children for point %d but found %d", i, actual)
			}
			if expected, actual := 1, len(children[0].items[0]); expected != actual {
				t.Fatalf("Expected one child item at level 0 for point %d but found %d", i, actual)
			}
		}
	})
}
