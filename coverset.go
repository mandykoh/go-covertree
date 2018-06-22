package covertree

type coverSet []coverSetItem

func coverSetWithItem(item, query Item) coverSet {
	return coverSet{coverSetItemForQuery(item, query)}
}

func (cs *coverSet) child(item Item, distThreshold float64, childLevel int, store Store) (child coverSet, err error) {
	for _, csItem := range *cs {

		if csItem.distance <= distThreshold {
			child = append(child, csItem)
		}

		children, err := store.Load(csItem.item, childLevel)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(children); i++ {
			childItem := coverSetItemForQuery(children[i], item)
			if childItem.distance <= distThreshold {
				child = append(child, childItem)
			}
		}
	}

	return
}

func (cs coverSet) minDistance() float64 {
	minDist := cs[0].distance

	for i := 1; i < len(cs); i++ {
		dist := cs[i].distance
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}
