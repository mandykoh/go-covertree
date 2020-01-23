package covertree

type coverSet []itemWithChildren

func coverSetWithItems(items []interface{}, parent interface{}, query interface{}, distanceFunc DistanceFunc, loadChildren func(...interface{}) ([]LevelsWithItems, error)) (coverSet, error) {
	var cs coverSet

	if len(items) > 0 {
		children, err := loadChildren(items...)
		if err != nil {
			return nil, err
		}

		for i, item := range items {
			distance := distanceFunc(item, query)
			iwc := itemWithChildren{withDistance: ItemWithDistance{item, distance}, parent: parent, children: children[i]}
			cs = append(cs, iwc)
		}
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

func (cs coverSet) child(query interface{}, distThreshold float64, childLevel int, distanceBetween DistanceFunc, loadChildren func(...interface{}) ([]LevelsWithItems, error)) (childCoverSet coverSet, parentWithinThreshold interface{}, err error) {
	childCoverSet = make(coverSet, 0, len(cs))
	promotedChildren := make(coverSet, 0, len(cs))
	promotedChildItems := make([]interface{}, 0, len(cs))

	for _, csItem := range cs {
		if csItem.withDistance.Distance <= distThreshold {
			childCoverSet = append(childCoverSet, csItem)
			parentWithinThreshold = csItem.withDistance.Item

			for _, childItem := range csItem.takeChildrenAt(childLevel) {
				if childDist := distanceBetween(childItem, query); childDist <= distThreshold {
					promotedChild := itemWithChildren{withDistance: ItemWithDistance{childItem, childDist}, parent: csItem.withDistance.Item}
					promotedChildren = append(promotedChildren, promotedChild)
					promotedChildItems = append(promotedChildItems, childItem)
				}
			}
		}
	}

	if len(promotedChildItems) > 0 {
		children, err := loadChildren(promotedChildItems...)
		if err != nil {
			return nil, nil, err
		}

		for i := range promotedChildItems {
			promotedChildren[i].children = children[i]
		}
		childCoverSet = append(childCoverSet, promotedChildren...)
	}

	return
}

func (cs coverSet) closest(maxItems int, maxDist float64) []ItemWithDistance {
	lastNonNil := 0
	mins := make([]ItemWithDistance, maxItems)

	for i := range cs {
		if cs[i].withDistance.Distance > maxDist {
			continue
		}

		for j := range mins {
			if mins[j].Item == nil || cs[i].withDistance.Distance < mins[j].Distance {
				for k := len(mins) - 1; k > j; k-- {
					mins[k] = mins[k-1]
				}
				lastNonNil++
				mins[j] = cs[i].withDistance
				break
			}
		}
	}

	if lastNonNil > len(mins) {
		lastNonNil = len(mins)
	}
	return mins[:lastNonNil]
}
