package covertree

type Store interface {
	LoadChildren(parent Item, level int) (children []Item, err error)
	LoadTree() (root Item, rootLevel, deepestLevel int, err error)
	SaveChild(child, parent Item, level int) error
	SaveTree(root Item, rootLevel, deepestLevel int) error
}
