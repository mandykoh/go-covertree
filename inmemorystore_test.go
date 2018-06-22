package covertree

import (
	"math"
	"testing"
)

type dummyItem struct {
	id    string
	value float64
}

func (item *dummyItem) Distance(other Item) float64 {
	otherItem := other.(*dummyItem)
	return math.Abs(item.value - otherItem.value)
}

func TestInMemoryStore(t *testing.T) {

	t.Run("Load()", func(t *testing.T) {
		parent := &dummyItem{"parent", 456.0}
		item1 := &dummyItem{"thing1", 123.0}
		item2 := &dummyItem{"thing2", 234.0}

		s := InMemoryStore{
			items: map[Item]map[int][]Item{
				parent: {
					7: {item1, item2},
				},
			},
		}

		t.Run("retrieves existing items", func(t *testing.T) {
			items, _ := s.Load(parent, 7)

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

			items, _ := s.Load(badParent, 7)

			if actual, expected := len(items), 0; actual != expected {
				t.Errorf("Expected %d items but found %d", expected, actual)
			}
		})

		t.Run("returns empty results for non-existent children", func(t *testing.T) {
			parent := &dummyItem{"parent", 456.0}

			items, _ := s.Load(parent, 5)

			if actual, expected := len(items), 0; actual != expected {
				t.Errorf("Expected %d items but found %d", expected, actual)
			}
		})
	})

	t.Run("Save()", func(t *testing.T) {

		t.Run("adds an entry for a new item", func(t *testing.T) {
			item := &dummyItem{"child", 123.0}
			parent := &dummyItem{"parent", 456.0}

			s := NewInMemoryStore()
			s.Save(item, parent, 5)

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

			s := NewInMemoryStore()
			s.Save(item1, parent, 5)

			item1.value = 234.0
			s.Save(item1, parent, 5)

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
	})
}
