package covertree

type itemWithChildren struct {
	parent   ItemWithDistance
	children ItemsWithLevels
}

func (iwc *itemWithChildren) hasChildren() bool {
	return len(iwc.children.items) > 0
}
