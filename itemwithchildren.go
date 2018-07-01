package covertree

type itemWithChildren struct {
	parent   ItemWithDistance
	children LevelsWithItems
}

func itemWithChildrenFromStore(item Item, distance float64, store Store) (iwc itemWithChildren, err error) {
	children, err := store.LoadChildren(item)
	if err != nil {
		return
	}

	return itemWithChildren{ItemWithDistance{item, distance}, children}, nil
}

func (iwc *itemWithChildren) hasChildren() bool {
	return len(iwc.children.items) > 0
}

func (iwc *itemWithChildren) removeChildrenAt(level int) []Item {
	return iwc.children.removeItemsAt(level)
}
