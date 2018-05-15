package covertree

type Store interface {
	Load(parent Item, level int) (items []Item, err error)
	Save(item, parent Item, level int) error
}
