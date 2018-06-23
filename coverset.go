package covertree

type coverSet []ItemWithDistance

func coverSetWithItem(item, query Item) coverSet {
	return coverSet{itemWithDistanceForQuery(item, query)}
}

func (cs coverSet) Len() int {
	return len(cs)
}

func (cs coverSet) Less(i, j int) bool {
	return cs[i].Distance <= cs[j].Distance
}

func (cs coverSet) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
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

func (cs coverSet) minDistance() float64 {
	minDist := cs[0].Distance

	for i := 1; i < len(cs); i++ {
		dist := cs[i].Distance
		if dist < minDist {
			minDist = dist
		}
	}

	return minDist
}
