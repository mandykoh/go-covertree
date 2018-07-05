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
	distanceBetween DistanceFunc
	store           Store
}

// NewEmptyTreeWithStore creates an empty Tree using the specified store.
//
// distanceFunc is the function used by the tree to determine the distance
// between two items.
func NewEmptyTreeWithStore(store Store, distanceFunc DistanceFunc) (*Tree, error) {
	return &Tree{
		root:            nil,
		rootLevel:       math.MaxInt32,
		distanceBetween: distanceFunc,
		store:           store,
	}, nil
}

// NewTreeFromStore creates and initialises a Tree from the specified store by
// calling the store’s LoadTree method.
//
// distanceFunc is the function used by the tree to determine the distance
// between two items.
func NewTreeFromStore(store Store, distanceFunc DistanceFunc) (*Tree, error) {
	root, rootLevel, err := store.LoadTree()
	if err != nil {
		return nil, err
	}

	return &Tree{
		root:            root,
		rootLevel:       rootLevel,
		distanceBetween: distanceFunc,
		store:           store,
	}, nil
}

// FindNearest returns the nearest items in the tree to the specified query
// item, up to the specified maximum number of results and maximum distance.
//
// Results are returned with their distances from the query Item, in order from
// closest to furthest.
//
// If no items are found matching the given criteria, an empty result set is
// returned.
func (t *Tree) FindNearest(query Item, maxResults int, maxDistance float64) (results []ItemWithDistance, err error) {
	if t.root == nil {
		return
	}

	cs, err := coverSetWithItem(t.root, t.distanceBetween(t.root, query), t.store)
	if err != nil {
		return nil, err
	}

	for level := t.rootLevel; !cs.atBottom(); level-- {
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
		if err != nil {
			return
		}
	}

	return cs.closest(maxResults, maxDistance), nil
}

// Insert attempts to insert the specified item into the tree.
//
// Two items are defined as being the same if their distance is exactly zero.
// If an item already exists in the tree which is the same as the item being
// inserted, the item in the tree is returned. Otherwise, the newly inserted
// item is returned instead.
func (t *Tree) Insert(item Item) (inserted Item, err error) {

	// Tree is empty - add item as the new root at infinity
	if t.root == nil {
		err := t.store.SaveTree(item, math.MaxInt32)
		if err == nil {
			t.root = item
			t.rootLevel = math.MaxInt32
		}
		return item, err
	}

	cs, err := coverSetWithItem(t.root, t.distanceBetween(t.root, item), t.store)
	if err != nil {
		return nil, err
	}

	// Tree only has a root at infinity - move root to appropriate level for the new item
	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(cs[0].parent.Distance)
		err = t.store.SaveTree(t.root, t.rootLevel)
		if err != nil {
			return
		}
	}

	inserted, err = t.insert(item, cs, t.rootLevel)

	if err == nil {
		if inserted == nil {

			// No covering parent found - promote the current root to cover the
			// new item and insert it as a child.

			err = t.hoistRootForChild(item, math.MinInt32)
			if err == nil {
				err = t.store.SaveTree(t.root, t.rootLevel)
			}
		}
	}

	return
}

// Remove removes the given item from the tree. If no such item exists in the
// tree, this has no effect.
func (t *Tree) Remove(item Item) (err error) {
	rootDist := t.distanceBetween(item, t.root)
	cs, err := coverSetWithItem(t.root, rootDist, t.store)
	if err != nil {
		return err
	}

	if rootDist == 0 {
		t.root = nil

		for level, items := range cs[0].children.items {
			for _, item := range items {
				if t.root == nil {

					// Promote one child to be the new root
					t.root = item

				} else {

					// Add all remaining children as children of the new root
					err = t.hoistRootForChild(item, level)
					if err != nil {
						return
					}
				}
			}
		}

		err = t.store.SaveTree(t.root, t.rootLevel)

	} else {
		oldRootLevel := t.rootLevel

		var orphans []Item
		orphans, err = t.remove(item, cs, t.rootLevel)

		for _, orphan := range orphans {
			err = t.hoistRootForChild(orphan, math.MinInt32)
			if err != nil {
				return err
			}
		}

		if t.rootLevel != oldRootLevel {
			err = t.store.SaveTree(t.root, t.rootLevel)
		}
	}

	return
}

func (t *Tree) adoptOrphans(orphans []Item, query Item, parents coverSet, distThreshold float64, childLevel int) ([]Item, error) {
	remaining := 0

nextOrphan:
	for _, item := range orphans {
		for _, parent := range parents {
			if parent.parent.Item != query && t.distanceBetween(item, parent.parent.Item) <= distThreshold {

				err := t.store.SaveChild(item, parent.parent.Item, childLevel)
				if err != nil {
					return nil, err
				}

				continue nextOrphan
			}
		}

		orphans[remaining] = item
		remaining++
	}

	return orphans[:remaining], nil
}

func (t *Tree) hoistRootForChild(child Item, minChildLevel int) (err error) {
	dist := t.distanceBetween(t.root, child)
	childLevel := levelForDistance(dist)
	newRootLevel := t.rootLevel

	if childLevel < minChildLevel {
		childLevel = minChildLevel
	}
	if childLevel >= newRootLevel {
		newRootLevel = childLevel + 1
	}

	err = t.store.SaveChild(child, t.root, childLevel)
	if err == nil {
		t.rootLevel = newRootLevel
	}
	return
}

func (t *Tree) insert(item Item, coverSet coverSet, level int) (inserted Item, err error) {
	distThreshold := distanceForLevel(level)

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, t.distanceBetween, t.store)
	if err != nil {
		return nil, err
	}

	if len(childCoverSet) > 0 {

		// Only one matching child which is at zero distance - item is a duplicate so return the original
		if childCoverSet[0].parent.Distance == 0 {
			return childCoverSet[0].parent.Item, nil
		}

		// Look for a suitable parent amongst the children
		inserted, err = t.insert(item, childCoverSet, level-1)
		if inserted != nil || err != nil {
			return
		}

		// No parent was found among the children - look for a suitable parent at this level
		for _, csItem := range coverSet {
			if csItem.parent.Distance <= distThreshold {
				err := t.store.SaveChild(item, csItem.parent.Item, level-1)
				return item, err
			}
		}
	}

	return nil, nil
}

func (t *Tree) remove(item Item, coverSet coverSet, level int) (orphans []Item, err error) {
	distThreshold := distanceForLevel(level)

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, t.distanceBetween, t.store)

	if err == nil && len(childCoverSet) > 0 {
		found := false

		for i := range childCoverSet {
			if childCoverSet[i].parent.Distance == 0 {
				found = true

				for _, csItem := range coverSet {
					err = t.store.DeleteChild(childCoverSet[i].parent.Item, csItem.parent.Item, level-1)
					if err != nil {
						return
					}
				}

				for _, child := range childCoverSet[i].children.items {
					orphans = append(orphans, child...)
				}

				// Try to get orphans adopted by one of the siblings of the deleted node
				orphans, err = t.adoptOrphans(orphans, item, childCoverSet, distanceForLevel(level-2), level-2)

				break
			}
		}

		if !found {
			orphans, err = t.remove(item, childCoverSet, level-1)
		}

		if err == nil {
			// Try to get orphans adopted by nodes at this level
			orphans, err = t.adoptOrphans(orphans, item, coverSet, distThreshold, level-1)
		}
	}

	return
}

func distanceForLevel(level int) float64 {
	return math.Pow(2, float64(level))
}

func levelForDistance(distance float64) int {
	return int(math.Ceil(math.Log2(distance)))
}
