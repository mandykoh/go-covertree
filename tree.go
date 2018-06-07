package covertree

import (
	"math"
)

type Tree struct {
	root      Item
	rootLevel int
}

func (t *Tree) Insert(item Item, store Store) (ok bool, err error) {
	if t.root == nil {
		t.root = item
		t.rootLevel = math.MaxInt32
		return true, nil
	}

	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = int(math.Log2(t.root.Distance(item)) + 1)
	}

	ok, err = insert(item, coverSet{t.root}, t.rootLevel, store)

	if !ok && err == nil {
		t.root, item = item, t.root
		t.rootLevel = int(math.Log2(t.root.Distance(item)) + 1)
		ok, err = insert(item, coverSet{t.root}, t.rootLevel, store)
	}

	return
}

func insert(item Item, coverSet coverSet, level int, store Store) (ok bool, err error) {
	distThreshold := math.Pow(2, float64(level))

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, store)
	if err != nil {
		return false, err
	}

	if len(childCoverSet) > 0 {
		ok, err = insert(item, childCoverSet, level-1, store)
		if ok || err != nil {
			return
		}

		for _, csItem := range coverSet {
			if item.Distance(csItem) <= distThreshold {
				err := store.Save(item, csItem, level-1)
				return err == nil, err
			}
		}
	}

	return false, nil
}
