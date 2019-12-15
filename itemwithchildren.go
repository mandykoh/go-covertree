package covertree

type itemWithChildren struct {
	withDistance ItemWithDistance
	parent       interface{}
	children     LevelsWithItems
}

func (iwc *itemWithChildren) hasChildren() bool {
	return len(iwc.children.items) > 0
}

func (iwc *itemWithChildren) takeChildrenAt(level int) []interface{} {
	return iwc.children.takeItemsAt(level)
}
