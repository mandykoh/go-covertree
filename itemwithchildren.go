package covertree

type itemWithChildren struct {
	withDistance ItemWithDistance
	parent       interface{}
	children     LevelsWithItems
}

func itemWithChildrenFromStore(item, parent interface{}, distance float64, store Store) (iwc itemWithChildren, err error) {
	children, err := store.LoadChildren(item)
	if err != nil {
		return
	}

	return itemWithChildren{ItemWithDistance{item, distance}, parent, children}, nil
}

func (iwc *itemWithChildren) hasChildren() bool {
	return len(iwc.children.items) > 0
}

func (iwc *itemWithChildren) takeChildrenAt(level int) []interface{} {
	return iwc.children.takeItemsAt(level)
}
