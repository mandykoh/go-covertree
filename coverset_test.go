package covertree

import (
	"math"
	"testing"
)

func TestCoverSet(t *testing.T) {

	t.Run("child()", func(t *testing.T) {

		expectResults := func(t *testing.T, actualResults, expectedResults coverSet) {

			if expected, actual := len(expectedResults), len(actualResults); expected != actual {
				t.Errorf("Expected %d results but found %d instead", expected, actual)
			}

			if len(expectedResults) < len(actualResults) {
				for i, c := range expectedResults {
					if expected, actual := c.parent.Item, actualResults[i].parent.Item; expected != actual {
						t.Errorf("Expected result %d to be %v but was %v", i, expected, actual)
					}
					if expected, actual := c.parent.Distance, actualResults[i].parent.Distance; expected != actual {
						t.Errorf("Expected result %d to have distance %g but was %g", i, expected, actual)
					}
				}
			}
		}

		t.Run("returns the child coverset which excludes non-covering items", func(t *testing.T) {
			cs := coverSet{
				{parent: ItemWithDistance{"a", 0.0}},
				{parent: ItemWithDistance{"b", 10.0}},
				{parent: ItemWithDistance{"c", 1.0}},
			}

			child, _ := cs.child("a", 2.0, 0, nil, nil)

			expectResults(t, child, coverSet{
				{parent: ItemWithDistance{"a", 0.0}},
				{parent: ItemWithDistance{"c", 1.0}},
			})
		})

		t.Run("promotes covering children at the requested level and excludes non-covering children", func(t *testing.T) {
			cs := coverSet{
				{parent: ItemWithDistance{"a", 0.0}, children: LevelsWithItems{items: map[int][]Item{3: {"c", "d"}}}},
				{parent: ItemWithDistance{"b", 10.0}},
			}

			mockDistFunc := func(a, b Item) float64 {
				if a == "c" || b == "c" {
					return 5.0
				}

				return 6.0
			}

			child, _ := cs.child("a", 5.0, 3, mockDistFunc, newInMemoryStore())

			expectResults(t, child, coverSet{
				{parent: ItemWithDistance{"a", 0.0}},
				{parent: ItemWithDistance{"c", 5.0}},
			})

			if cs[0].hasChildren() {
				t.Errorf("Expected promoted children to no longer be children")
			}
		})
	})

	t.Run("closest()", func(t *testing.T) {

		expectResults := func(t *testing.T, actualResults, expectedResults []ItemWithDistance) {

			if expected, actual := len(expectedResults), len(actualResults); expected != actual {
				t.Errorf("Expected %d results but found %d instead", expected, actual)
			}

			if len(expectedResults) < len(actualResults) {
				for i, c := range expectedResults {
					if expected, actual := c.Item, actualResults[i].Item; expected != actual {
						t.Errorf("Expected result %d to be %v but was %v", i, expected, actual)
					}
					if expected, actual := c.Distance, actualResults[i].Distance; expected != actual {
						t.Errorf("Expected result %d to have distance %g but was %g", i, expected, actual)
					}
				}
			}
		}

		t.Run("returns the specified number of items from closest to furthest", func(t *testing.T) {
			cs := coverSet{
				{parent: ItemWithDistance{"a", 5.0}},
				{parent: ItemWithDistance{"c", 3.0}},
				{parent: ItemWithDistance{"b", 4.0}},
				{parent: ItemWithDistance{"e", 1.0}},
				{parent: ItemWithDistance{"d", 2.0}},
			}

			results := cs.closest(3, math.MaxFloat64)

			expectResults(t, results, []ItemWithDistance{
				{"e", 1.0},
				{"d", 2.0},
				{"c", 3.0},
			})
		})

		t.Run("returns all available results up to the number requested", func(t *testing.T) {
			cs := coverSet{
				{parent: ItemWithDistance{"a", 5.0}},
				{parent: ItemWithDistance{"c", 3.0}},
				{parent: ItemWithDistance{"b", 4.0}},
			}

			results := cs.closest(4, math.MaxFloat64)

			expectResults(t, results, []ItemWithDistance{
				{"c", 3.0},
				{"b", 4.0},
				{"a", 5.0},
			})
		})

		t.Run("returns all available results up to the distance limit", func(t *testing.T) {
			cs := coverSet{
				{parent: ItemWithDistance{"a", 5.0}},
				{parent: ItemWithDistance{"c", 3.0}},
				{parent: ItemWithDistance{"b", 4.0}},
			}

			results := cs.closest(3, 4.0)

			expectResults(t, results, []ItemWithDistance{
				{"c", 3.0},
				{"b", 4.0},
			})
		})
	})
}