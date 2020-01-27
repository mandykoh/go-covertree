package covertree

import (
	"math/rand"
	"sort"
	"testing"
)

func TestCoverSetLayer(t *testing.T) {

	t.Run("constrainedToDistance()", func(t *testing.T) {

		t.Run("returns a new coverSetLayer containing only items within the specified distance", func(t *testing.T) {
			items := []itemWithChildren{
				{
					withDistance: ItemWithDistance{
						Item:     "item1",
						Distance: 0.0,
					},
				},
				{
					withDistance: ItemWithDistance{
						Item:     "item2",
						Distance: 1.0,
					},
				},
				{
					withDistance: ItemWithDistance{
						Item:     "item3",
						Distance: 2.0,
					},
				},
			}

			layer := makeCoverSetLayer(items)
			result := layer.constrainedToDistance(1.0)

			if expected, actual := 2, len(result); expected != actual {
				t.Fatalf("Expected constrained layer to contain %d items but found %d", expected, actual)
			}

			for i := range result {
				if expected, actual := items[i], result[i]; expected.withDistance.Item != actual.withDistance.Item {
					t.Errorf("Expected '%s' with distance %v to be in position %d but found %s with distance %v instead", expected.withDistance.Item, expected.withDistance.Distance, i, actual.withDistance.Item, actual.withDistance.Distance)
				}
			}

			result = layer.constrainedToDistance(1.5)

			if expected, actual := 2, len(result); expected != actual {
				t.Fatalf("Expected constrained layer to contain %d items but found %d", expected, actual)
			}

			for i := range result {
				if expected, actual := items[i], result[i]; expected.withDistance.Item != actual.withDistance.Item {
					t.Errorf("Expected '%s' with distance %v to be in position %d but found %s with distance %v instead", expected.withDistance.Item, expected.withDistance.Distance, i, actual.withDistance.Item, actual.withDistance.Distance)
				}
			}

			result = layer.constrainedToDistance(0.99)

			if expected, actual := 1, len(result); expected != actual {
				t.Fatalf("Expected constrained layer to contain %d items but found %d", expected, actual)
			}

			for i := range result {
				if expected, actual := items[i], result[i]; expected.withDistance.Item != actual.withDistance.Item {
					t.Errorf("Expected '%s' with distance %v to be in position %d but found %s with distance %v instead", expected.withDistance.Item, expected.withDistance.Distance, i, actual.withDistance.Item, actual.withDistance.Distance)
				}
			}
		})
	})

	t.Run("makeCoverSetLayer()", func(t *testing.T) {

		t.Run("creates layer with items in sorted order by distance", func(t *testing.T) {
			items := []itemWithChildren{
				{
					withDistance: ItemWithDistance{
						Item:     "item1",
						Distance: rand.Float64(),
					},
				},
				{
					withDistance: ItemWithDistance{
						Item:     "item2",
						Distance: rand.Float64(),
					},
				},
				{
					withDistance: ItemWithDistance{
						Item:     "item3",
						Distance: rand.Float64(),
					},
				},
			}

			rand.Shuffle(len(items), func(i, j int) {
				items[i], items[j] = items[j], items[i]
			})

			layer := makeCoverSetLayer(items)

			itemsSorted := make([]itemWithChildren, len(items))
			copy(itemsSorted, items)
			sort.Slice(itemsSorted, func(i, j int) bool {
				return itemsSorted[i].withDistance.Distance < itemsSorted[j].withDistance.Distance
			})

			if expected, actual := len(items), len(layer); expected != actual {
				t.Fatalf("Expected cover set layer to contain %d items but found %d", expected, actual)
			}

			for i, expected := range itemsSorted {
				if actual := layer[i]; expected.withDistance.Item != actual.withDistance.Item {
					t.Errorf("Expected '%s' with distance %v to be in position %d but found %s with distance %v instead", expected.withDistance.Item, expected.withDistance.Distance, i, actual.withDistance.Item, actual.withDistance.Distance)
				}
			}
		})
	})
}
