package covertree

import (
	"math"
	"sync"
)

// Tree represents a single cover tree.
//
// Trees should generally not be created except via NewTreeFromStore, and then
// only by a Store.
type Tree struct {
	basis           float64
	distanceBetween DistanceFunc
	store           Store
	rootMutex       sync.RWMutex
}

// NewTreeWithStore creates and initialises a Tree using the specified store.
//
// basis is the logarithmic base for determining the coverage of nodes at each
// level of the tree.
//
// distanceFunc is the function used by the tree to determine the distance
// between two items.
func NewTreeWithStore(store Store, basis float64, distanceFunc DistanceFunc) (*Tree, error) {
	return &Tree{
		basis:           basis,
		distanceBetween: distanceFunc,
		store:           store,
	}, nil
}

// FindNearest returns the nearest items in the tree to the specified query
// item, up to the specified maximum number of results and maximum distance.
//
// Results are returned with their distances from the query item, in order from
// closest to furthest.
//
// If no items are found matching the given criteria, an empty result set is
// returned.
//
// Multiple calls to FindNearest and Insert are safe to make concurrently.
func (t *Tree) FindNearest(query interface{}, maxResults int, maxDistance float64) (results []ItemWithDistance, err error) {
	return t.findNearestWithTrace(query, maxResults, maxDistance, t.NewTracer())
}

// Insert inserts the specified item into the tree.
//
// Multiple calls to FindNearest and Insert are safe to make concurrently.
func (t *Tree) Insert(item interface{}) (err error) {
	return t.insertWithTrace(item, t.NewTracer())
}

// NewTracer returns a new Tracer for recording performance metrics for
// operations on this tree.
//
// Tracers are not thread safe and should not be shared by multiple Goroutines.
func (t *Tree) NewTracer() *Tracer {
	return &Tracer{tree: t}
}

// Remove removes the given item from the tree. If no such item exists in the
// tree, this has no effect.
//
// removed will be the item that was successfully removed, or nil if no matching
// item was found.
//
// This method is not safe for concurrent use. Calls to Remove should be
// externally synchronised so they do not execute concurrently with each other
// or with calls to FindNearest or Insert.
func (t *Tree) Remove(item interface{}) (removed interface{}, err error) {
	return t.removeWithTrace(item, t.NewTracer())
}

func (t *Tree) adoptOrphans(orphans []interface{}, query interface{}, parents coverSet, distThreshold float64, childLevel int) ([]interface{}, error) {
	remaining := 0

nextOrphan:
	for _, item := range orphans {
		for _, parent := range parents {
			if parent.withDistance.Item != query && t.distanceBetween(item, parent.withDistance.Item) <= distThreshold {

				err := t.store.UpdateItem(item, parent.withDistance.Item, childLevel)
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

func (t *Tree) distanceForLevel(level int) float64 {
	return math.Pow(t.basis, float64(level))
}

func (t *Tree) findNearestWithTrace(query interface{}, maxResults int, maxDistance float64, tracer *Tracer) (results []ItemWithDistance, err error) {
	var root interface{}
	var rootLevel int

	err = t.withRootReadLock(func() error {
		root, rootLevel, err = t.loadRoot()
		return err
	})
	if err != nil || root == nil {
		return
	}

	cs, err := coverSetWithItem(root, nil, t.distanceBetween(root, query), t.store.LoadChildren)
	if err != nil {
		return nil, err
	}

	tracer.recordLevel(cs)

	for level := rootLevel; !cs.atBottom(); level-- {
		distThreshold := t.distanceForLevel(level)

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

		cs, _, err = cs.child(query, distThreshold, level-1, t.distanceBetween, t.store.LoadChildren)
		if err != nil {
			return
		}

		tracer.recordLevel(cs)
	}

	return cs.closest(maxResults, maxDistance), nil
}

func (t *Tree) hoistRootForChild(child interface{}, minChildLevel int, root interface{}, rootLevel int) (newRootLevel int, err error) {
	dist := t.distanceBetween(root, child)
	childLevel := t.levelForDistance(dist)
	newRootLevel = rootLevel

	if childLevel < minChildLevel {
		childLevel = minChildLevel
	}
	if childLevel >= newRootLevel {
		newRootLevel = childLevel + 1
	}

	err = t.store.UpdateItem(child, root, childLevel)
	return
}

func (t *Tree) insert(item interface{}, coverSet coverSet, level int, tracer *Tracer) (inserted interface{}, err error) {
	distThreshold := t.distanceForLevel(level)

	childCoverSet, parentWithinThreshold, err := coverSet.child(item, distThreshold, level-1, t.distanceBetween, t.store.LoadChildren)
	if err != nil || len(childCoverSet) == 0 {
		return nil, err
	}

	tracer.recordLevel(childCoverSet)

	// A matching child which is at zero distance - item is a duplicate so insert it as a child
	if childCoverSet[0].withDistance.Distance == 0 {
		err = t.store.AddItem(item, childCoverSet[0].withDistance.Item, level-2)
		return item, err
	}

	// Look for a suitable parent amongst the children
	inserted, err = t.insert(item, childCoverSet, level-1, tracer)
	if inserted != nil || err != nil {
		return
	}

	// No parent was found among the children - pick arbitrary suitable parent at this level
	if parentWithinThreshold != nil {
		err = t.store.AddItem(item, parentWithinThreshold, level-1)
		return item, err
	}

	return nil, nil
}

func (t *Tree) insertWithTrace(item interface{}, tracer *Tracer) (err error) {
	var root interface{}
	var rootLevel int
	var cs coverSet

	err = t.withRootReadLock(func() error {
		root, rootLevel, err = t.loadRoot()
		if err != nil {
			return err
		}

		if root != nil {
			cs, err = coverSetWithItem(root, nil, t.distanceBetween(root, item), t.store.LoadChildren)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return
	}

	tracer.recordLevel(cs)

	if root == nil || rootLevel == math.MaxInt32 {
		var newRoot bool

		err = t.withRootWriteLock(func() error {
			root, rootLevel, err = t.loadRoot()
			if err != nil {
				return err
			}

			// Tree is empty - add item as the new root at infinity
			if root == nil {
				newRoot = true
				err = t.store.AddItem(item, nil, math.MaxInt32)
				return err
			}

			cs, err = coverSetWithItem(root, nil, t.distanceBetween(root, item), t.store.LoadChildren)
			if err != nil {
				return err
			}

			// Tree only has a root at infinity - move root to appropriate level for the new item
			if rootLevel == math.MaxInt32 {
				rootLevel = t.levelForDistance(cs[0].withDistance.Distance)
				return t.store.UpdateItem(root, nil, rootLevel)
			}

			return nil
		})
		if err != nil || newRoot {
			return
		}
	}

	var inserted interface{}
	inserted, err = t.insert(item, cs, rootLevel, tracer)

	if err == nil {
		if inserted == nil {

			// No covering parent found - promote the current root to cover the
			// new item and insert it as a child.
			err = t.withRootWriteLock(func() error {
				root, rootLevel, err = t.loadRoot()
				if err != nil {
					return err
				}

				rootLevel, err = t.hoistRootForChild(item, math.MinInt32, root, rootLevel)
				if err == nil {
					return t.store.UpdateItem(root, nil, rootLevel)
				}

				return nil
			})
		}
	}

	return
}

func (t *Tree) levelForDistance(distance float64) int {
	return int(math.Ceil(math.Log2(distance) / math.Log2(t.basis)))
}

func (t *Tree) loadRoot() (root interface{}, rootLevel int, err error) {
	rootLevels, err := t.store.LoadChildren(nil)
	if err != nil {
		return
	}

	for level, items := range rootLevels.items {
		root = items[0]
		rootLevel = level
		break
	}

	return
}

func (t *Tree) remove(item interface{}, coverSet coverSet, level int, tracer *Tracer) (removed interface{}, orphans []interface{}, err error) {
	distThreshold := t.distanceForLevel(level)

	childCoverSet, _, err := coverSet.child(item, distThreshold, level-1, t.distanceBetween, t.store.LoadChildren)

	if err == nil {
		tracer.recordLevel(childCoverSet)

		if len(childCoverSet) > 0 {
			for i := range childCoverSet {
				if childCoverSet[i].withDistance.Distance == 0 {
					err = t.store.RemoveItem(childCoverSet[i].withDistance.Item, childCoverSet[i].parent, level-1)
					if err != nil {
						return
					}
					removed = childCoverSet[i].withDistance.Item

					for _, child := range childCoverSet[i].children.items {
						orphans = append(orphans, child...)
					}

					// Try to get orphans adopted by one of the siblings of the deleted node
					orphans, err = t.adoptOrphans(orphans, item, childCoverSet, t.distanceForLevel(level-2), level-2)

					break
				}
			}

			if removed == nil {
				removed, orphans, err = t.remove(item, childCoverSet, level-1, tracer)
			}

			if err == nil {
				// Try to get orphans adopted by nodes at this level
				orphans, err = t.adoptOrphans(orphans, item, coverSet, distThreshold, level-1)
			}
		}
	}

	return
}

func (t *Tree) removeWithTrace(item interface{}, tracer *Tracer) (removed interface{}, err error) {
	root, rootLevel, err := t.loadRoot()
	if err != nil || root == nil {
		return nil, err
	}

	rootDist := t.distanceBetween(item, root)
	cs, err := coverSetWithItem(root, nil, rootDist, t.store.LoadChildren)
	if err != nil {
		return nil, err
	}

	tracer.recordLevel(cs)

	if rootDist == 0 {
		err = t.store.RemoveItem(root, nil, rootLevel)
		if err != nil {
			return
		}

		removed = root
		root = nil

		for level, items := range cs[0].children.items {
			for _, item := range items {
				if root == nil {

					// Promote one child to be the new root
					root = item

				} else {

					// Add all remaining children as children of the new root
					rootLevel, err = t.hoistRootForChild(item, level, root, rootLevel)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		err = t.store.UpdateItem(root, nil, rootLevel)

	} else {
		var orphans []interface{}
		removed, orphans, err = t.remove(item, cs, rootLevel, tracer)

		if err == nil {
			oldRootLevel := rootLevel

			for _, orphan := range orphans {
				rootLevel, err = t.hoistRootForChild(orphan, math.MinInt32, root, rootLevel)
				if err != nil {
					return nil, err
				}
			}

			if rootLevel != oldRootLevel {
				err = t.store.UpdateItem(root, nil, rootLevel)
			}
		}
	}

	return
}

func (t *Tree) withRootReadLock(f func() error) error {
	t.rootMutex.RLock()
	defer t.rootMutex.RUnlock()
	return t.store.WithRootReadLock(f)
}

func (t *Tree) withRootWriteLock(f func() error) error {
	t.rootMutex.Lock()
	defer t.rootMutex.Unlock()
	return t.store.WithRootWriteLock(f)
}
