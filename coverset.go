package covertree

type coverSet []itemWithChildren

func coverSetWithItem(item, parent interface{}, distance float64, store Store) (coverSet, error) {
	iwc, err := itemWithChildrenFromStore(item, parent, distance, store)
	if err != nil {
		return nil, err
	}

	return coverSet{iwc}, nil
}

func (cs coverSet) atBottom() bool {
	for _, csItem := range cs {
		if csItem.hasChildren() {
			return false
		}
	}
	return true
}

func (cs coverSet) child(query interface{}, distThreshold float64, childLevel int, distanceBetween DistanceFunc, store Store) (childCoverSet coverSet, parentWithinThreshold interface{}, err error) {
	childCoverSet = make(coverSet, 0, len(cs))

	for _, csItem := range cs {
		if csItem.withDistance.Distance <= distThreshold {
			childCoverSet = append(childCoverSet, csItem)
			parentWithinThreshold = csItem.withDistance.Item
		}

		for _, childItem := range csItem.takeChildrenAt(childLevel) {
			if childDist := distanceBetween(childItem, query); childDist <= distThreshold {
				promotedChild, err := itemWithChildrenFromStore(childItem, csItem.withDistance.Item, childDist, store)
				if err != nil {
					return nil, nil, err
				}

				childCoverSet = append(childCoverSet, promotedChild)
			}
		}
	}

	return
}

func (cs coverSet) closest(maxItems int, maxDist float64) []ItemWithDistance {
	mins := make([]ItemWithDistance, maxItems, maxItems)

	for i := range cs {
		if cs[i].withDistance.Distance > maxDist {
			continue
		}

		for j := range mins {
			if mins[j].Item == nil || cs[i].withDistance.Distance < mins[j].Distance {
				for k := len(mins) - 1; k > j; k-- {
					mins[k] = mins[k-1]
				}
				mins[j] = cs[i].withDistance
				break
			}
		}
	}

	lastNonNil := len(mins) - 1
	for lastNonNil >= 0 && mins[lastNonNil].Item == nil {
		lastNonNil--
	}
	return mins[:lastNonNil+1]
}
