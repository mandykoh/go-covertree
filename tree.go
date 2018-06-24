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
			distThreshold += closest[len(closest)-1].Distance
		}

		cs, err = cs.child(query, distThreshold, level-1, store)
		if err != nil || len(cs) == 0 {
			return
		}
	}

	return cs.closest(maxResults, maxDistance), nil
}

func (t *Tree) Insert(item Item, store Store) error {

	// Tree is empty - add item as the new root at infinity
	if t.root == nil {
		t.root = item
		t.rootLevel = math.MaxInt32
		t.deepestLevel = t.rootLevel
		return nil
	}

	cs := coverSetWithItem(t.root, item)

	// Tree only has a root at infinity - move root to appropriate level for the new item
	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(cs[0].Distance)
		t.deepestLevel = t.rootLevel
	}

	parentFoundAtLevel, err := insert(item, cs, t.rootLevel, store)

	if err == nil {

		if parentFoundAtLevel < math.MaxInt32 {

			// A parent was found - update the tree depth if appropriate
			if parentFoundAtLevel < t.deepestLevel {
				t.deepestLevel = parentFoundAtLevel - 1
			}

		} else {

			// No parent found - re-parent the tree with the new item
			cs := coverSetWithItem(item, t.root)
			newRootLevel := levelForDistance(cs[0].Distance)

			_, err = insert(t.root, cs, newRootLevel, store)
			if err == nil {
				t.root = item
				t.rootLevel = newRootLevel
			}
		}
	}

	return err
}

func distanceForLevel(level int) float64 {
	return math.Pow(2, float64(level))
}

func insert(item Item, coverSet coverSet, level int, store Store) (parentFoundAtLevel int, err error) {
	distThreshold := distanceForLevel(level)

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, store)
	if err != nil {
		return math.MaxInt32, err
	}

	if len(childCoverSet) > 0 {

		// Look for a suitable parent amongst the children
		parentFoundAtLevel, err = insert(item, childCoverSet, level-1, store)
		if parentFoundAtLevel < math.MaxInt32 || err != nil {
			return
		}

		// No parent was found among the children - look for a suitable parent at this level
		for _, csItem := range coverSet {
			if csItem.Distance <= distThreshold {
				err := store.Save(item, csItem.Item, level-1)
				return level, err
			}
		}
	}

	return math.MaxInt32, nil
}

func levelForDistance(distance float64) int {
	return int(math.Log2(distance) + 1)
}
