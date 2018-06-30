package covertree

import (
	"math"
)

// Tree represents a single cover tree.
//
// Trees should generally not be created except via NewTreeFromStore, and then
// only by a Store.
type Tree struct {
	root            Item
	rootLevel       int
	deepestLevel    int
	distanceBetween DistanceFunc
	store           Store
}

// NewTreeFromStore creates and initialises a Tree from the specified store by
// calling the store’s LoadTree method. Depending on the store, this may result
// in a new empty tree being created, or a previously populated tree being
// returned.
//
// distanceFunc is the function used by the tree to determine the distance
// between two items.
func NewTreeFromStore(store Store, distanceFunc DistanceFunc) (*Tree, error) {
	root, rootLevel, deepestLevel, err := store.LoadTree()
	if err != nil {
		return nil, err
	}

	return &Tree{
		root:            root,
		rootLevel:       rootLevel,
		deepestLevel:    deepestLevel,
		distanceBetween: distanceFunc,
		store:           store,
	}, nil
}

func (t *Tree) FindNearest(query Item, maxResults int, maxDistance float64) (results []ItemWithDistance, err error) {
	if t.root == nil {
		return
	}

	cs := coverSetWithItem(t.root, t.distanceBetween(t.root, query))

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

		cs, err = cs.child(query, distThreshold, level-1, t.distanceBetween, t.store)
		if err != nil || len(cs) == 0 {
			return
		}
	}

	return cs.closest(maxResults, maxDistance), nil
}

func (t *Tree) Insert(item Item) (inserted Item, err error) {

	// Tree is empty - add item as the new root at infinity
	if t.root == nil {
		err := t.store.SaveTree(item, math.MaxInt32, math.MaxInt32)
		if err == nil {
			t.root = item
			t.rootLevel = math.MaxInt32
			t.deepestLevel = t.rootLevel
		}
		return item, err
	}

	cs := coverSetWithItem(t.root, t.distanceBetween(t.root, item))

	// Tree only has a root at infinity - move root to appropriate level for the new item
	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(cs[0].Distance)
	}

	inserted, insertedLevel, err := t.insert(item, cs, t.rootLevel)

	if err == nil {
		if inserted == nil {

			// No parent found - re-parent the tree with the new item
			cs := coverSetWithItem(item, t.distanceBetween(item, t.root))
			newRootLevel := levelForDistance(cs[0].Distance)

			inserted, insertedLevel, err = t.insert(t.root, cs, newRootLevel)
			if err == nil {
				err = t.store.SaveTree(item, newRootLevel, t.deepestLevel)
				if err == nil {
					t.root = item
					t.rootLevel = newRootLevel
				}
			}

		} else if insertedLevel < t.deepestLevel {
			err = t.store.SaveTree(t.root, t.rootLevel, insertedLevel)
			if err == nil {
				t.deepestLevel = insertedLevel
			}
		}
	}

	return
}

func (t *Tree) insert(item Item, coverSet coverSet, level int) (inserted Item, insertedLevel int, err error) {
	distThreshold := distanceForLevel(level)

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, t.distanceBetween, t.store)
	if err != nil {
		return nil, level, err
	}

	if len(childCoverSet) > 0 {

		// Only one matching child which is at zero distance - item is a duplicate so return the original
		if childCoverSet[0].Distance == 0 {
			return childCoverSet[0].Item, level - 1, nil
		}

		// Look for a suitable parent amongst the children
		inserted, insertedLevel, err = t.insert(item, childCoverSet, level-1)
		if inserted != nil || err != nil {
			return
		}

		// No parent was found among the children - look for a suitable parent at this level
		for _, csItem := range coverSet {
			if csItem.Distance <= distThreshold {
				err := t.store.SaveChild(item, csItem.Item, level-1)
				return item, level - 1, err
			}
		}
	}

	return nil, level, nil
}

func distanceForLevel(level int) float64 {
	return math.Pow(2, float64(level))
}

func levelForDistance(distance float64) int {
	return int(math.Log2(distance) + 1)
}
