package covertree

import (
	"math"
)

// Tree represents a single cover tree.
//
// Trees should generally not be created except via NewTreeFromStore, and then
// only by a Store.
type Tree struct {
	basis           float64
	rootLevel       int
	distanceBetween DistanceFunc
	store           Store
}

// NewTreeWithStore creates and initialises a Tree using the specified store.
//
// basis is the logarithmic base for determining the coverage of nodes at each
// level of the tree.
//
// rootDistance is the minimum expected distance between root nodes. New nodes
// that exceed this distance will be created as additional roots.
//
// distanceFunc is the function used by the tree to determine the distance
// between two items.
func NewTreeWithStore(store Store, basis float64, rootDistance float64, distanceFunc DistanceFunc) (*Tree, error) {
	tree := &Tree{
		basis:           basis,
		distanceBetween: distanceFunc,
		store:           store,
	}
	tree.rootLevel = tree.levelForDistance(rootDistance)

	return tree, nil
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
	cs, err := t.loadRootCoverSet(query, tracer)
	if err != nil {
		return nil, err
	}

	tracer.recordLevel(cs)

	for level := t.rootLevel; !cs.atBottom(); level-- {
		distThreshold := t.distanceForLevel(level)

		if closest := cs.closest(maxResults, maxDistance); len(closest) == maxResults {
			distThreshold += closest[len(closest)-1].Distance
		} else {
			distThreshold += maxDistance
		}

		cs, _, err = cs.child(query, distThreshold, level-1, t.distanceBetween, tracer.loadChildren)
		if err != nil {
			return
		}

		tracer.recordLevel(cs)
	}

	return cs.closest(maxResults, maxDistance), nil
}

func (t *Tree) hoistRootForChild(child interface{}, minChildLevel int, root interface{}, rootLevel int) (newRootLevel, newChildLevel int) {
	dist := t.distanceBetween(root, child)
	childLevel := t.levelForDistance(dist)
	newRootLevel = rootLevel

	if childLevel < minChildLevel {
		childLevel = minChildLevel
	}
	if childLevel >= newRootLevel {
		newRootLevel = childLevel + 1
	}

	return newRootLevel, childLevel
}

func (t *Tree) insert(item interface{}, coverSet coverSet, level int, tracer *Tracer) (inserted interface{}, err error) {
	distThreshold := t.distanceForLevel(level)

	childCoverSet, parentWithinThreshold, err := coverSet.child(item, distThreshold, level-1, t.distanceBetween, tracer.loadChildren)
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

func (t *Tree) insertWithTrace(item interface{}, tracer *Tracer) error {
	cs, err := t.loadRootCoverSet(item, tracer)
	if err != nil {
		return err
	}

	tracer.recordLevel(cs)

	var inserted interface{}
	inserted, err = t.insert(item, cs, t.rootLevel, tracer)
	if err == nil && inserted == nil {
		return t.store.AddItem(item, nil, t.rootLevel)
	}

	return err
}

func (t *Tree) levelForDistance(distance float64) int {
	return int(math.Ceil(math.Log2(distance) / math.Log2(t.basis)))
}

func (t *Tree) loadRootCoverSet(query interface{}, tracer *Tracer) (coverSet, error) {
	roots, err := tracer.loadChildren(nil)
	if err != nil {
		return nil, err
	}

	return coverSetWithItems(roots[0].itemsAt(t.rootLevel), nil, query, t.distanceBetween, tracer.loadChildren)
}

func (t *Tree) remove(item interface{}, coverSet coverSet, level int, tracer *Tracer) (removed interface{}, orphans []interface{}, err error) {
	for i := range coverSet {
		if coverSet[i].withDistance.Distance == 0 {
			err = t.store.RemoveItem(coverSet[i].withDistance.Item, coverSet[i].parent, level)
			if err != nil {
				return
			}
			removed = coverSet[i].withDistance.Item

			for _, child := range coverSet[i].children.items {
				orphans = append(orphans, child...)
			}

			// Try to get orphans adopted by one of the siblings of the deleted node
			orphans, err = t.adoptOrphans(orphans, item, coverSet, t.distanceForLevel(level-1), level-1)

			break
		}
	}

	if removed == nil {
		if coverSet.atBottom() {
			return
		}

		distThreshold := t.distanceForLevel(level)
		childCoverSet, _, err := coverSet.child(item, distThreshold, level-1, t.distanceBetween, tracer.loadChildren)
		if err != nil {
			return nil, nil, err
		}
		tracer.recordLevel(childCoverSet)

		removed, orphans, err = t.remove(item, childCoverSet, level-1, tracer)
	}

	return
}

func (t *Tree) removeWithTrace(item interface{}, tracer *Tracer) (removed interface{}, err error) {
	cs, err := t.loadRootCoverSet(item, tracer)
	if err != nil {
		return nil, err
	}

	tracer.recordLevel(cs)

	var orphans []interface{}
	removed, orphans, err = t.remove(item, cs, t.rootLevel, tracer)
	if err == nil {
		for _, orphan := range orphans {
			err = t.store.UpdateItem(orphan, nil, t.rootLevel)
			if err != nil {
				return nil, err
			}
		}
	}

	return
}
