package covertree

type Store interface {
	LoadChildren(parent Item, level int) (children []Item, err error)
	SaveChild(child, parent Item, level int) error
}
