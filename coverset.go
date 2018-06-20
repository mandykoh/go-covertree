package covertree

type coverSet []coverSetItem

func (cs *coverSet) child(item Item, distThreshold float64, childLevel int, store Store) (child coverSet, err error) {
	for _, csItem := range *cs {

		if csItem.Distance(item) <= distThreshold {
			child = append(child, csItem)
		}

		children, err := store.Load(csItem.item, childLevel)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(children); i++ {
			childItem := makeCoverSetItem(children[i])
			if childItem.Distance(item) <= distThreshold {
				child = append(child, childItem)
			}
		}
	}

	return
}
