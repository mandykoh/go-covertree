package covertree

type coverSet []itemWithChildren

func coverSetWithItem(item Item, distance float64, store Store) (coverSet, error) {
	children, err := store.LoadChildren(item)
	if err != nil {
		return nil, err
	}

	return coverSet{{ItemWithDistance{item, distance}, children}}, nil
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

			for _, childItem := range csItem.children.itemsForLevel(childLevel) {
				childItem := ItemWithDistance{childItem, distanceBetween(childItem, query)}
				if childItem.Distance <= distThreshold {
					childItemChildren, err := store.LoadChildren(childItem.Item)
					if err != nil {
						return nil, err
					}

					promotedChild := itemWithChildren{childItem, childItemChildren}
					child = append(child, promotedChild)
				}
			}

			csItem.children.deleteLevel(childLevel)
		}
	}

	return
}

func (cs coverSet) closest(maxItems int, maxDist float64) []ItemWithDistance {
	mins := make([]ItemWithDistance, maxItems, maxItems)

	for i := 0; i < len(cs); i++ {
		if cs[i].parent.Distance > maxDist {
			continue
		}

		for j := 0; j < len(mins); j++ {
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
