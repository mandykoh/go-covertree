package covertree

type coverSet []itemWithChildren

func coverSetWithItem(item Item, distance float64, store Store) (coverSet, error) {
	iwc, err := itemWithChildrenFromStore(item, distance, store)
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

func (cs coverSet) child(query Item, distThreshold float64, childLevel int, distanceBetween DistanceFunc, store Store) (child coverSet, err error) {
	for _, csItem := range cs {
		if csItem.parent.Distance <= distThreshold {
			child = append(child, csItem)
		}

		for _, childItem := range csItem.removeChildrenAt(childLevel) {
			childDist := distanceBetween(childItem, query)
			if childDist <= distThreshold {
				promotedChild, err := itemWithChildrenFromStore(childItem, childDist, store)
				if err != nil {
					return nil, err
				}

				child = append(child, promotedChild)
			}
		}
	}

	return
}

func (cs coverSet) closest(maxItems int, maxDist float64) []ItemWithDistance {
	mins := make([]ItemWithDistance, maxItems, maxItems)

	for i := range cs {
		if cs[i].parent.Distance > maxDist {
			continue
		}

		for j := range mins {
			if mins[j].Item == nil || cs[i].parent.Distance < mins[j].Distance {
				for k := len(mins) - 1; k > j; k-- {
					mins[k] = mins[k-1]
				}
				mins[j] = cs[i].parent
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
