package covertree

type coverSet []ItemWithDistance

func coverSetWithItem(item, query Item) coverSet {
	return coverSet{itemWithDistanceForQuery(item, query)}
}

func (cs coverSet) child(item Item, distThreshold float64, childLevel int, store Store) (child coverSet, err error) {
	for _, csItem := range cs {

		if csItem.Distance <= distThreshold {
			child = append(child, csItem)
		}

		children, err := store.Load(csItem.Item, childLevel)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(children); i++ {
			childItem := itemWithDistanceForQuery(children[i], item)
			if childItem.Distance <= distThreshold {
				child = append(child, childItem)
			}
		}
	}

	return
}

func (cs coverSet) closest(maxItems int, maxDist float64) []ItemWithDistance {
	mins := make([]ItemWithDistance, maxItems, maxItems)

	for i := 0; i < len(cs); i++ {
		if cs[i].Distance > maxDist {
			continue
		}

		for j := 0; j < len(mins); j++ {
			if mins[j].Item == nil || cs[i].Distance < mins[j].Distance {
				for k := len(mins) - 1; k > j; k-- {
					mins[k] = mins[k-1]
				}
				mins[j] = cs[i]
				break
			}
		}
	}

	for i := len(mins) - 1; i >= 0; i-- {
		if mins[i].Item != nil {
			return mins[:i+1]
		}
	}

	return nil
}
