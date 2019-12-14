package covertree

type coverSet []itemWithChildren

func coverSetWithItems(items []interface{}, parent interface{}, query interface{}, distanceFunc DistanceFunc, loadChildren func(interface{}) (LevelsWithItems, error)) (coverSet, error) {
	var cs coverSet

	for _, item := range items {
		distance := distanceFunc(item, query)
		iwc, err := itemWithChildrenFromStore(item, parent, distance, loadChildren)
		if err != nil {
			return nil, err
		}

		cs = append(cs, iwc)
	}

	return cs, nil
}

func (cs coverSet) atBottom() bool {
	for _, csItem := range cs {
		if csItem.hasChildren() {
			return false
		}
	}
	return true
}

func (cs coverSet) child(query interface{}, distThreshold float64, childLevel int, distanceBetween DistanceFunc, loadChildren func(interface{}) (LevelsWithItems, error)) (childCoverSet coverSet, parentWithinThreshold interface{}, err error) {
	childCoverSet = make(coverSet, 0, len(cs))

	for _, csItem := range cs {
		if csItem.withDistance.Distance <= distThreshold {
			childCoverSet = append(childCoverSet, csItem)
			parentWithinThreshold = csItem.withDistance.Item

			for _, childItem := range csItem.takeChildrenAt(childLevel) {
				if childDist := distanceBetween(childItem, query); childDist <= distThreshold {
					promotedChild, err := itemWithChildrenFromStore(childItem, csItem.withDistance.Item, childDist, loadChildren)
					if err != nil {
						return nil, nil, err
					}

					childCoverSet = append(childCoverSet, promotedChild)
				}
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
