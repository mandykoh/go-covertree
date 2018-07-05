package covertree

import (
	"math"
	"testing"
)

type dummyItem struct {
	id    string
	value float64
}

func TestInMemoryStore(t *testing.T) {

	distanceBetween := func(a, b Item) float64 {
		return math.Abs(a.(*dummyItem).value - b.(*dummyItem).value)
	}

	t.Run("DeleteChild()", func(t *testing.T) {
		parent := &dummyItem{"parent", 456.0}
		item1 := &dummyItem{"thing1", 123.0}
		item2 := &dummyItem{"thing2", 234.0}

		setup := func() inMemoryStore {
			return inMemoryStore{
				distanceBetween: distanceBetween,
				items: map[Item]map[int][]Item{
					parent: {
						7: {item1, item2},
					},
				},
			}
		}

		t.Run("removes an existing child", func(t *testing.T) {
			s := setup()
			s.DeleteChild(item2, parent, 7)

			items := s.levelsFor(parent)[7]

			if expected, actual := 1, len(items); expected != actual {
				t.Errorf("Expected one child item after deletion but got %d", actual)
			}
			if expected, actual := item1, items[0]; expected != actual {
				t.Errorf("Expected remaining child to be %v but was %v", expected, actual)
			}
		})

		t.Run("removes an existing child by distance", func(t *testing.T) {
			s := setup()
			s.DeleteChild(&dummyItem{"something", 234.0}, parent, 7)

			items := s.levelsFor(parent)[7]

			if expected, actual := 1, len(items); expected != actual {
				t.Errorf("Expected one child item after deletion but got %d", actual)
			}
			if expected, actual := item1, items[0]; expected != actual {
				t.Errorf("Expected remaining child to be %v but was %v", expected, actual)
			}
		})
	})

	t.Run("LoadChildren()", func(t *testing.T) {
		parent := &dummyItem{"parent", 456.0}
		item1 := &dummyItem{"thing1", 123.0}
		item2 := &dummyItem{"thing2", 234.0}

		s := inMemoryStore{
			items: map[Item]map[int][]Item{
				parent: {
					7: {item1, item2},
				},
			},
		}

		t.Run("retrieves existing child items", func(t *testing.T) {
			allChildren, _ := s.LoadChildren(parent)
			items := allChildren.itemsAt(7)

			if actual, expected := len(items), 2; actual != expected {
				t.Errorf("Expected %d items but found %d", expected, actual)
			} else {
				if actual, expected := items[0], item1; actual != expected {
					t.Errorf("Expected item '%s' but found %v", expected.id, actual)
				}
				if actual, expected := items[1], item2; actual != expected {
					t.Errorf("Expected item '%s' but found %v", expected.id, actual)
				}
			}
		})

		t.Run("returns empty results for non-existent parent", func(t *testing.T) {
			badParent := &dummyItem{"bad parent", 456.0}

			allChildren, _ := s.LoadChildren(badParent)

			if actual, expected := len(allChildren.items), 0; actual != expected {
				t.Errorf("Expected %d items but found %d", expected, actual)
			}
		})

		t.Run("returns empty results for non-existent children", func(t *testing.T) {
			parent := &dummyItem{"parent", 456.0}

			allChildren, _ := s.LoadChildren(parent)
			items := allChildren.itemsAt(5)

			if actual, expected := len(items), 0; actual != expected {
				t.Errorf("Expected %d items but found %d", expected, actual)
			}
		})
	})

	t.Run("SaveChild()", func(t *testing.T) {

		t.Run("adds an entry for a new child item", func(t *testing.T) {
			item := &dummyItem{"child", 123.0}
			parent := &dummyItem{"parent", 456.0}

			s := newInMemoryStore(nil)
			s.SaveChild(item, parent, 5)

			levels, ok := s.items[item]
			if !ok {
				t.Fatalf("Expected levels to exist for new item %v", item)
			}

			levels, ok = s.items[parent]
			if !ok {
				t.Fatalf("Expected levels to exist for parent %v", parent)
			}
			if actual, expected := len(levels[5]), 1; actual != expected {
				t.Errorf("Expected 1 item in level but found %d", actual)
			}
			if actual, expected := levels[5][0].(*dummyItem), item; actual != expected {
				t.Errorf("Expected item %f but found %f", expected.value, actual.value)
			}
		})

		t.Run("overwrites an entry for an existing item", func(t *testing.T) {
			item1 := &dummyItem{"child", 123.0}
			parent := &dummyItem{"parent", 456.0}

			s := newInMemoryStore(nil)
			s.SaveChild(item1, parent, 5)

			item1.value = 234.0
			s.SaveChild(item1, parent, 5)

			levels, ok := s.items[parent]
			if !ok {
				t.Fatalf("Expected levels to exist for parent %v", parent)
			}
			if actual, expected := len(levels[5]), 1; actual != expected {
				t.Errorf("Expected 1 item in level but found %d", actual)
			}
			if actual, expected := levels[5][0].(*dummyItem).value, 234.0; actual != expected {
				t.Errorf("Expected item %f but found %f", expected, actual)
			}
		})

		t.Run("overwrites an entry for an existing item by distance", func(t *testing.T) {
			item1 := &dummyItem{"child", 123.0}
			parent := &dummyItem{"parent", 456.0}

			s := newInMemoryStore(distanceBetween)
			s.SaveChild(item1, parent, 5)

			s.SaveChild(&dummyItem{"new value", 123.0}, parent, 5)

			levels, ok := s.items[parent]
			if !ok {
				t.Fatalf("Expected levels to exist for parent %v", parent)
			}
			if actual, expected := len(levels[5]), 1; actual != expected {
				t.Errorf("Expected 1 item in level but found %d", actual)
			}
			if actual, expected := levels[5][0].(*dummyItem).id, "new value"; actual != expected {
				t.Errorf("Expected item %v but found %v", expected, actual)
			}
		})
	})
}
