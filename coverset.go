package covertree

type coverSet []Item

func (cs *coverSet) child(item Item, distThreshold float64, childLevel int, store Store) (child coverSet, err error) {
	for _, csItem := range *cs {

		if item.Distance(csItem) <= distThreshold {
			child = append(child, csItem)
		}

		children, err := store.Load(csItem, childLevel)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(children); i++ {
			childItem := children[i]
			if item.Distance(childItem) <= distThreshold {
				child = append(child, childItem)
			}
		}
	}

	return
}
