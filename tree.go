package covertree

import (
	"math"
)

type Tree struct {
	root         Item
	rootLevel    int
	deepestLevel int
}

func (t *Tree) FindNearest(query Item, store Store, maxResults int, maxDistance float64) (results []ItemWithDistance, err error) {
	cs := coverSetWithItem(t.root, query)

	for level := t.rootLevel; level >= t.deepestLevel; level-- {
		distThreshold := distanceForLevel(level)

		closest := cs.closest(maxResults, maxDistance)
		if len(closest) > 0 {
			closestDist := closest[len(closest)-1].Distance
			if len(closest) < maxResults {
				distThreshold += math.Max(maxDistance, closestDist)
			} else {
				distThreshold += closestDist
			}
		} else {
			distThreshold += maxDistance
		}

		cs, err = cs.child(query, distThreshold, level-1, store)
		if err != nil || len(cs) == 0 {
			return
		}
	}

	return cs.closest(maxResults, maxDistance), nil
}

func (t *Tree) Insert(item Item, store Store) (inserted Item, err error) {

	// Tree is empty - add item as the new root at infinity
	if t.root == nil {
		t.root = item
		t.rootLevel = math.MaxInt32
		t.deepestLevel = t.rootLevel
		return item, nil
	}

	cs := coverSetWithItem(t.root, item)

	// Tree only has a root at infinity - move root to appropriate level for the new item
	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(cs[0].Distance)
		t.deepestLevel = t.rootLevel
	}

	inserted, err = t.insert(item, cs, t.rootLevel, store)

	if err == nil && inserted == nil {

		// No parent found - re-parent the tree with the new item
		cs := coverSetWithItem(item, t.root)
		newRootLevel := levelForDistance(cs[0].Distance)

		inserted, err = t.insert(t.root, cs, newRootLevel, store)
		if err == nil {
			t.root = item
			t.rootLevel = newRootLevel
		}
	}

	return
}

func (t *Tree) insert(item Item, coverSet coverSet, level int, store Store) (inserted Item, err error) {
	distThreshold := distanceForLevel(level)

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, store)
	if err != nil {
		return nil, err
	}

	if len(childCoverSet) > 0 {

		// Only one matching child which is at zero distance - item is a duplicate so return the original
		if childCoverSet[0].Distance == 0 {
			return childCoverSet[0].Item, nil
		}

		// Look for a suitable parent amongst the children
		inserted, err = t.insert(item, childCoverSet, level-1, store)
		if inserted != nil || err != nil {
			return
		}

		// No parent was found among the children - look for a suitable parent at this level
		for _, csItem := range coverSet {
			if csItem.Distance <= distThreshold {
				err := store.Save(item, csItem.Item, level-1)
				if err == nil && level-1 < t.deepestLevel {
					t.deepestLevel = level - 1
				}

				return item, err
			}
		}
	}

	return nil, nil
}

func distanceForLevel(level int) float64 {
	return math.Pow(2, float64(level))
}

func levelForDistance(distance float64) int {
	return int(math.Log2(distance) + 1)
}
