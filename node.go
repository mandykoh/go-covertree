package covertree

import (
	"math"
)

func Insert(item Item, coverSet coverSet, level int, store Store) (ok bool, err error) {
	distThreshold := math.Pow(2, float64(level))

	childCoverSet, err := coverSet.child(item, distThreshold, level-1, store)
	if err != nil {
		return false, err
	}

	if len(childCoverSet) > 0 {
		ok, err = Insert(item, childCoverSet, level-1, store)
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
