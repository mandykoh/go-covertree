package covertree

import (
	"github.com/mandykoh/go-parallel"
	"sync/atomic"
)

// CompositeTree spreads operations across multiple subtrees for scaling and
// parallelisation.
type CompositeTree struct {
	trees       []*Tree
	insertCount uint32
}

// FindNearest returns the nearest items in all the subtrees to the specified
// query item, up to the specified maximum number of results and maximum
// distance.
//
// Subtrees are queried in parallel.
//
// Results are returned with their distances from the query item, in order from
// closest to furthest.
//
// If no items are found matching the given criteria, an empty result set is
// returned.
//
// Multiple calls to FindNearest and Insert are safe to make concurrently.
func (ct *CompositeTree) FindNearest(query interface{}, maxResults int, maxDistance float64) (results []ItemWithDistance, err error) {

	subResults := make([][]ItemWithDistance, len(ct.trees))
	var subErr error

	parallel.RunWorkers(len(subResults), func(workerNum, workerCount int) {
		results, err := ct.trees[workerNum].FindNearest(query, maxResults, maxDistance)
		if err != nil {
			subErr = err
			return
		}

		subResults[workerNum] = results
	})
	if subErr != nil {
		return nil, subErr
	}

	return zipItemsWithDistance(subResults, maxResults), nil
}

// Insert inserts the specified item into one of the subtrees.
//
// Multiple calls to FindNearest and Insert are safe to make concurrently.
func (ct *CompositeTree) Insert(item interface{}) (err error) {
	treeIndex := atomic.AddUint32(&ct.insertCount, 1) % uint32(len(ct.trees))
	return ct.trees[treeIndex].Insert(item)
}

func NewCompositeTree(trees ...*Tree) *CompositeTree {
	return &CompositeTree{
		trees: trees,
	}
}

func zipItemsWithDistance(itemSets [][]ItemWithDistance, limit int) []ItemWithDistance {
	var results []ItemWithDistance

	itemIndices := make([]int, len(itemSets))

	for itemCount := 0; itemCount < limit; itemCount++ {

		// Find the minimum item across all item sets
		minItemSet := 0
		var minItem *ItemWithDistance

		for i := range itemSets {
			itemIndex := itemIndices[i]
			if itemIndex >= len(itemSets[i]) {
				continue
			}

			item := &itemSets[i][itemIndex]
			if minItem == nil || item.Distance < minItem.Distance {
				minItem = item
				minItemSet = i
			}
		}

		// No minimum item found - no more items, so stop
		if minItem == nil {
			break
		}

		itemIndices[minItemSet]++
		results = append(results, *minItem)
	}

	return results
}
