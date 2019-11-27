package covertree

import (
	"fmt"
	"time"
)

// Tracer represents a record for performance metrics of Tree operations.
//
// Tracers for a given tree can be created using the tree’s NewTracer method.
//
// Tracers are not thread safe and should not be shared by multiple Goroutines.
type Tracer struct {
	tree               *Tree
	MaxCoverSetSize    int
	MaxLevelsTraversed int
	LoadChildrenCount  int
	TotalTime          time.Duration
}

// FindNearest returns the nearest items in the tree to the specified query
// item, up to the specified maximum number of results and maximum distance.
//
// Results are returned with their distances from the query item, in order from
// closest to furthest.
//
// If no items are found matching the given criteria, an empty result set is
// returned.
func (t *Tracer) FindNearest(query interface{}, maxResults int, maxDistance float64) (results []ItemWithDistance, err error) {
	t.doWithTrace(func() {
		results, err = t.tree.findNearestWithTrace(query, maxResults, maxDistance, t)
	})
	return
}

// Insert inserts the specified item into the tree.
func (t *Tracer) Insert(item interface{}) (err error) {
	t.doWithTrace(func() {
		err = t.tree.insertWithTrace(item, t)
	})
	return
}

// Remove removes the given item from the tree. If no such item exists in the
// tree, this has no effect.
//
// removed will be the item that was successfully removed, or nil if no matching
// item was found.
func (t *Tracer) Remove(item interface{}) (removed interface{}, err error) {
	t.doWithTrace(func() {
		removed, err = t.tree.removeWithTrace(item, t)
	})
	return
}

func (t *Tracer) String() string {
	return fmt.Sprintf("%v, cover set size: %d, levels traversed: %d, load children count: %d", t.TotalTime, t.MaxCoverSetSize, t.MaxLevelsTraversed, t.LoadChildrenCount)
}

func (t *Tracer) doWithTrace(f func()) {
	var startTime time.Time

	t.reset()

	defer func() {
		t.TotalTime = time.Now().Sub(startTime)
	}()

	startTime = time.Now()
	f()
}

func (t *Tracer) loadChildren(parent interface{}) (LevelsWithItems, error) {
	t.LoadChildrenCount++
	return t.tree.store.LoadChildren(parent)
}

func (t *Tracer) recordLevel(cs coverSet) {
	size := len(cs)
	if size > t.MaxCoverSetSize {
		t.MaxCoverSetSize = size
	}

	t.MaxLevelsTraversed++
}

func (t *Tracer) reset() {
	t.MaxCoverSetSize = 0
	t.MaxLevelsTraversed = 0
	t.LoadChildrenCount = 0
}
