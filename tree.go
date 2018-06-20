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
	parentCoverSet := coverSet{makeCoverSetItem(t.root, query)}

	var childCoverSet coverSet
	for level >= t.deepestLevel {
		distThreshold := parentCoverSet[0].distance
		for i := 1; i < len(parentCoverSet); i++ {
			dist := parentCoverSet[i].distance
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

	for i := range childCoverSet {
		results = append(results, childCoverSet[i].item)
	}
	return
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

	parentFound, insertLevel, err := insert(item, coverSet{makeCoverSetItem(t.root, item)}, t.rootLevel, store)

	if err == nil {
		if parentFound {
			if insertLevel < t.deepestLevel {
				t.deepestLevel = insertLevel
			}

		} else {
			fmt.Println("Reparenting")
			newRootLevel := levelForDistance(item, t.root)

			_, _, err = insert(t.root, coverSet{makeCoverSetItem(item, t.root)}, newRootLevel, store)
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
			if csItem.distance <= distThreshold {
				err := store.Save(item, csItem.item, level-1)
				return err == nil, level - 1, err
			}
		}
	}

	return false, level, nil
}

func levelForDistance(item1, item2 Item) int {
	return int(math.Log2(item1.Distance(item2)) + 1)
}
