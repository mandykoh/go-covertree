package covertree

import (
	"math"
	"sync"
)

type Tree struct {
	root         Item
	rootLevel    int
	deepestLevel int
	mutex        sync.Mutex
}

func (t *Tree) FindNearest(query Item, store Store) (results []Item, err error) {
	cs := coverSetWithItem(t.root, query)

	for level := t.rootLevel; level >= t.deepestLevel; level-- {
		distThreshold := cs.minDistance() + distanceForLevel(level)

		cs, err = cs.child(query, distThreshold, level-1, store)
		if err != nil {
			return
		}
	}

	for i := range cs {
		results = append(results, cs[i].item)
	}
	return
}

func (t *Tree) Insert(item Item, store Store) error {
	t.mutex.Lock()

	// Tree is empty - add item as the new root at infinity
	if t.root == nil {
		t.root = item
		t.rootLevel = math.MaxInt32
		t.deepestLevel = t.rootLevel
		t.mutex.Unlock()
		return nil
	}

	cs := coverSetWithItem(t.root, item)

	// Tree only has a root at infinity - move root to appropriate level for the new item
	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(cs[0].distance)
		t.deepestLevel = t.rootLevel
	}

	t.mutex.Unlock()

	parentFoundAtLevel, err := insert(item, cs, t.rootLevel, store)

	if err == nil {
		t.mutex.Lock()

		if parentFoundAtLevel < math.MaxInt32 {

			// A parent was found - update the tree depth if appropriate
			if parentFoundAtLevel < t.deepestLevel {
				t.deepestLevel = parentFoundAtLevel - 1
			}

		} else {

			// No parent found - re-parent the tree with the new item
			cs := coverSetWithItem(item, t.root)
			newRootLevel := levelForDistance(cs[0].distance)

			_, err = insert(t.root, cs, newRootLevel, store)
			if err == nil {
				t.root = item
				t.rootLevel = newRootLevel
			}
		}

		t.mutex.Unlock()
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
			if csItem.distance <= distThreshold {
				err := store.Save(item, csItem.item, level-1)
				return level, err
			}
		}
	}

	return math.MaxInt32, nil
}

func levelForDistance(distance float64) int {
	return int(math.Log2(distance) + 1)
}
