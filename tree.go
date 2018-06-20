package covertree

import (
	"fmt"
	"math"
)

type Tree struct {
	root         Item
	rootLevel    int
	deepestLevel int
}

func (t *Tree) FindNearest(query Item, store Store) (results []Item, err error) {
	level := t.rootLevel
	parentCoverSet := coverSet{t.root}

	var childCoverSet coverSet
	for level >= t.deepestLevel {
		distThreshold := query.Distance(parentCoverSet[0])
		for i := 1; i < len(parentCoverSet); i++ {
			dist := query.Distance(parentCoverSet[i])
			if dist < distThreshold {
				distThreshold = dist
			}
		}
		distThreshold += math.Pow(2, float64(level))

		level -= 1
		childCoverSet, err = parentCoverSet.child(query, distThreshold, level, store)
		if err != nil {
			return
		}

		parentCoverSet = childCoverSet
	}

	return childCoverSet, nil
}

func (t *Tree) Insert(item Item, store Store) error {
	if t.root == nil {
		t.root = item
		t.rootLevel = math.MaxInt32
		t.deepestLevel = t.rootLevel
		return nil
	}

	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(t.root, item)
		t.deepestLevel = t.rootLevel
	}

	parentFound, insertLevel, err := insert(item, coverSet{t.root}, t.rootLevel, store)

	if err == nil {
		if parentFound {
			if insertLevel < t.deepestLevel {
				t.deepestLevel = insertLevel
			}

		} else {
			fmt.Println("Reparenting")
			newRootLevel := levelForDistance(item, t.root)

			_, _, err = insert(t.root, coverSet{item}, newRootLevel, store)
			if err == nil {
				t.root = item
				t.rootLevel = newRootLevel
			}
		}
	}

	return err
}

func insert(item Item, coverSet coverSet, level int, store Store) (parentFound bool, insertLevel int, err error) {
	distThreshold := math.Pow(2, float64(level))

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, store)
	if err != nil {
		return false, level, err
	}

	if len(childCoverSet) > 0 {
		parentFound, insertLevel, err = insert(item, childCoverSet, level-1, store)
		if parentFound || err != nil {
			return
		}

		for _, csItem := range coverSet {
			if item.Distance(csItem) <= distThreshold {
				err := store.Save(item, csItem, level-1)
				return err == nil, level - 1, err
			}
		}
	}

	return false, level, nil
}

func levelForDistance(item1, item2 Item) int {
	return int(math.Log2(item1.Distance(item2)) + 1)
}
