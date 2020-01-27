package covertree

import (
	"math"
	"testing"
)

func TestCoverSet(t *testing.T) {

	t.Run("child()", func(t *testing.T) {

		expectResults := func(t *testing.T, actualResults, expectedResults coverSet) {
			t.Helper()

			if expected, actual := len(expectedResults.layers), len(actualResults.layers); expected != actual {
				t.Fatalf("Expected %d layers in cover set but found %d instead", expected, actual)
			}

			if expected, actual := expectedResults.itemCount, actualResults.itemCount; expected != actual {
				t.Errorf("Expected cover set to have item count of %d but was %d", expected, actual)
			}

			for layerNum := 0; layerNum < len(expectedResults.layers); layerNum++ {
				expectedLayer := expectedResults.layers[layerNum]
				actualLayer := actualResults.layers[layerNum]

				if expected, actual := len(expectedLayer), len(actualLayer); expected != actual {
					t.Errorf("Expected layer %d to contain %d items but found %d instead", layerNum, expected, actual)

				} else {
					for itemNum := range expectedLayer {
						expectedItem := expectedLayer[itemNum]
						actualItem := actualLayer[itemNum]

						if expected, actual := expectedItem.withDistance.Item, actualItem.withDistance.Item; expected != actual {
							t.Errorf("Expected item %d of layer %d to be %v but was %v", itemNum, layerNum, expected, actual)
						}
						if expected, actual := expectedItem.withDistance.Distance, actualItem.withDistance.Distance; expected != actual {
							t.Errorf("Expected item %d of layer %d to have distance %g but was %g", itemNum, layerNum, expected, actual)
						}
					}
				}
			}
		}

		t.Run("returns the child coverset which excludes non-covering items", func(t *testing.T) {
			var cs coverSet
			cs.addLayer(makeCoverSetLayer([]itemWithChildren{
				{withDistance: ItemWithDistance{"a", 0.0}},
				{withDistance: ItemWithDistance{"b", 10.0}},
				{withDistance: ItemWithDistance{"c", 1.0}},
			}))

			child, _, _ := cs.child("a", 2.0, 0, nil, nil)

			var expectedCoverSet coverSet
			expectedCoverSet.addLayer(makeCoverSetLayer([]itemWithChildren{
				{withDistance: ItemWithDistance{"a", 0.0}},
				{withDistance: ItemWithDistance{"c", 1.0}},
			}))

			expectResults(t, child, expectedCoverSet)
		})

		t.Run("promotes covering children at the requested level and excludes non-covering children", func(t *testing.T) {
			var cs coverSet
			cs.addLayer(makeCoverSetLayer([]itemWithChildren{
				{withDistance: ItemWithDistance{"a", 0.0}, children: LevelsWithItems{items: map[int][]interface{}{3: {"c", "d"}}}},
				{withDistance: ItemWithDistance{"b", 10.0}},
			}))

			mockDistFunc := func(a, b interface{}) float64 {
				if a == "c" || b == "c" {
					return 5.0
				}

				return 6.0
			}

			store := NewInMemoryStore(nil)
			child, _, _ := cs.child("a", 5.0, 3, mockDistFunc, store.LoadChildren)

			var expectedCoverSet coverSet
			expectedCoverSet.addLayer(makeCoverSetLayer([]itemWithChildren{
				cs.layers[0][0],
			}))
			expectedCoverSet.addLayer(makeCoverSetLayer([]itemWithChildren{
				{withDistance: ItemWithDistance{"c", 5.0}},
			}))

			expectResults(t, child, expectedCoverSet)
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
				layers: []coverSetLayer{
					makeCoverSetLayer([]itemWithChildren{
						{withDistance: ItemWithDistance{"a", 5.0}},
						{withDistance: ItemWithDistance{"c", 3.0}},
						{withDistance: ItemWithDistance{"b", 4.0}},
						{withDistance: ItemWithDistance{"e", 1.0}},
						{withDistance: ItemWithDistance{"d", 2.0}},
					}),
				},
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
				layers: []coverSetLayer{
					makeCoverSetLayer([]itemWithChildren{
						{withDistance: ItemWithDistance{"a", 5.0}},
						{withDistance: ItemWithDistance{"c", 3.0}},
						{withDistance: ItemWithDistance{"b", 4.0}},
					}),
				},
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
				layers: []coverSetLayer{
					makeCoverSetLayer([]itemWithChildren{
						{withDistance: ItemWithDistance{"a", 5.0}},
						{withDistance: ItemWithDistance{"c", 3.0}},
						{withDistance: ItemWithDistance{"b", 4.0}},
					}),
				},
			}

			results := cs.closest(3, 4.0)

			expectResults(t, results, []ItemWithDistance{
				{"c", 3.0},
				{"b", 4.0},
			})
		})
	})
}
