package covertree

import (
	"math"
)

type Tree struct {
	root      Item
	rootLevel int
}

func (t *Tree) Insert(item Item, store Store) error {
	if t.root == nil {
		t.root = item
		t.rootLevel = math.MaxInt32
		return nil
	}

	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(t.root, item)
	}

	parentFound, err := insert(item, coverSet{t.root}, t.rootLevel, store)

	if !parentFound && err == nil {
		t.root, item = item, t.root
		t.rootLevel = levelForDistance(t.root, item)
		_, err = insert(item, coverSet{t.root}, t.rootLevel, store)
	}

	return err
}

func insert(item Item, coverSet coverSet, level int, store Store) (parentFound bool, err error) {
	distThreshold := math.Pow(2, float64(level))

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, store)
	if err != nil {
		return false, err
	}

	if len(childCoverSet) > 0 {
		parentFound, err = insert(item, childCoverSet, level-1, store)
		if parentFound || err != nil {
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

func levelForDistance(item1, item2 Item) int {
	return int(math.Log2(item1.Distance(item2)) + 1)
}
