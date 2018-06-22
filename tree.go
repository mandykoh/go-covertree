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
		distThreshold := cs.minDistance() + math.Pow(2, float64(level))

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
	if t.root == nil {
		t.root = item
		t.rootLevel = math.MaxInt32
		t.deepestLevel = t.rootLevel
		t.mutex.Unlock()
		return nil
	}

	cs := coverSetWithItem(t.root, item)

	if t.rootLevel == math.MaxInt32 {
		t.rootLevel = levelForDistance(cs[0].distance)
		t.deepestLevel = t.rootLevel
	}
	t.mutex.Unlock()

	parentFoundAtLevel, err := insert(item, cs, t.rootLevel, store)

	if err == nil {
		t.mutex.Lock()

		if parentFoundAtLevel < math.MaxInt32 {
			if parentFoundAtLevel < t.deepestLevel {
				t.deepestLevel = parentFoundAtLevel - 1
			}

		} else {
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

func insert(item Item, coverSet coverSet, level int, store Store) (parentFoundAtLevel int, err error) {
	distThreshold := math.Pow(2, float64(level))

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, store)
	if err != nil {
		return math.MaxInt32, err
	}

	if len(childCoverSet) > 0 {
		parentFoundAtLevel, err = insert(item, childCoverSet, level-1, store)
		if parentFoundAtLevel < math.MaxInt32 || err != nil {
			return
		}

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
