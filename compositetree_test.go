package covertree

import (
	"math"
	"testing"
)

func TestCompositeTree(t *testing.T) {

	isPointInResults := func(p *Point, results []ItemWithDistance) bool {
		for i := range results {
			if results[i].Item == p {
				return true
			}
		}

		return false
	}

	t.Run("FindNearest()", func(t *testing.T) {

		t.Run("searches across all subtrees", func(t *testing.T) {

			ct := NewCompositeTree(
				NewInMemoryTree(2, 1000.0, distanceBetweenPoints),
				NewInMemoryTree(2, 1000.0, distanceBetweenPoints),
			)

			points := randomPoints(4)
			_ = ct.trees[0].Insert(&points[0])
			_ = ct.trees[0].Insert(&points[1])
			_ = ct.trees[1].Insert(&points[2])
			_ = ct.trees[1].Insert(&points[3])

			p := randomPoint()
			results, err := ct.FindNearest(&p, 8, math.MaxFloat64)
			if err != nil {
				t.Fatalf("Expected successful find but got error: %v", err)
			}

			for i := range points {
				if !isPointInResults(&points[i], results) {
					t.Errorf("Expected to find point %v in results but did not", points[i])
				}
			}
		})
	})

	t.Run("Insert()", func(t *testing.T) {

		t.Run("distributes items across subtrees", func(t *testing.T) {

			ct := NewCompositeTree(
				NewInMemoryTree(2, 1000.0, distanceBetweenPoints),
				NewInMemoryTree(2, 1000.0, distanceBetweenPoints),
			)

			points := randomPoints(4)
			for i := range points {
				err := ct.Insert(&points[i])
				if err != nil {
					t.Fatalf("Expected successful insert but got error: %v", err)
				}
			}

			results0, _ := ct.trees[0].FindNearest(&Point{}, 8, math.MaxFloat64)
			results1, _ := ct.trees[1].FindNearest(&Point{}, 8, math.MaxFloat64)

			if isPointInResults(&points[0], results0) {
				t.Errorf("Expected not to find %v in tree 0 but did", points[0])
			}
			if !isPointInResults(&points[1], results0) {
				t.Errorf("Expected to find %v in tree 0 but did not", points[1])
			}
			if isPointInResults(&points[2], results0) {
				t.Errorf("Expected not to find %v in tree 0 but did", points[2])
			}
			if !isPointInResults(&points[3], results0) {
				t.Errorf("Expected to find %v in tree 0 but did not", points[3])
			}

			if !isPointInResults(&points[0], results1) {
				t.Errorf("Expected to find %v in tree 1 but did not", points[0])
			}
			if isPointInResults(&points[1], results1) {
				t.Errorf("Expected not to find %v in tree 1 but did", points[1])
			}
			if !isPointInResults(&points[2], results1) {
				t.Errorf("Expected to find %v in tree 1 but did not", points[2])
			}
			if isPointInResults(&points[3], results1) {
				t.Errorf("Expected not to find %v in tree 1 but did", points[3])
			}
		})
	})

	t.Run("zipItemsWithDistance()", func(t *testing.T) {

		assertResults := func(t *testing.T, expected, actual []ItemWithDistance) {
			t.Helper()

			if expectedLen, actualLen := len(expected), len(actual); expectedLen != actualLen {
				t.Errorf("Expected %d items but got %d", expectedLen, actualLen)
			} else {
				for i := 0; i < len(expected); i++ {
					if expected[i].Item != actual[i].Item {
						t.Fatalf("Expected %v but got %v", expected, actual)
					}
				}
			}
		}

		t.Run("picks results across item sets", func(t *testing.T) {
			itemSets := [][]ItemWithDistance{
				{
					{Item: "a", Distance: 0},
					{Item: "b", Distance: 1},
				},
				{
					{Item: "c", Distance: 0.5},
					{Item: "d", Distance: 1.5},
				},
			}

			results := zipItemsWithDistance(itemSets, 2)

			assertResults(t,
				[]ItemWithDistance{
					{Item: "a", Distance: 0},
					{Item: "c", Distance: 0.5},
				},
				results)
		})

		t.Run("picks results with lowest distances", func(t *testing.T) {
			itemSets := [][]ItemWithDistance{
				{
					{Item: "a", Distance: 1},
					{Item: "b", Distance: 2},
				},
				{
					{Item: "c", Distance: 0},
					{Item: "d", Distance: 3},
				},
			}

			results := zipItemsWithDistance(itemSets, 2)

			assertResults(t,
				[]ItemWithDistance{
					{Item: "c", Distance: 0},
					{Item: "a", Distance: 1},
				},
				results)
		})

		t.Run("picks results from other item sets when item sets run out", func(t *testing.T) {
			itemSets := [][]ItemWithDistance{
				{
					{Item: "a", Distance: 0},
					{Item: "b", Distance: 0},
				},
				{
					{Item: "c", Distance: 0},
					{Item: "d", Distance: 0},
				},
			}

			results := zipItemsWithDistance(itemSets, 3)

			assertResults(t,
				[]ItemWithDistance{
					{Item: "a", Distance: 0},
					{Item: "b", Distance: 0},
					{Item: "c", Distance: 0},
				},
				results)
		})

		t.Run("picks up to the available number of results when more are requested", func(t *testing.T) {
			itemSets := [][]ItemWithDistance{
				{
					{Item: "a", Distance: 0},
					{Item: "b", Distance: 2},
				},
				{
					{Item: "c", Distance: 0},
					{Item: "d", Distance: 2},
				},
			}

			results := zipItemsWithDistance(itemSets, 5)

			assertResults(t,
				[]ItemWithDistance{
					{Item: "a", Distance: 0},
					{Item: "c", Distance: 0},
					{Item: "b", Distance: 2},
					{Item: "d", Distance: 2},
				},
				results)
		})
	})
}
