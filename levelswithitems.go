package covertree

// LevelsWithItems represents a set of child Items of a parent Item, separated
// into their levels.
type LevelsWithItems struct {
	items map[int][]Item
}

// Add adds an Item to the specified level.
func (lwi *LevelsWithItems) Add(level int, item Item) {
	if lwi.items == nil {
		lwi.items = make(map[int][]Item)
	}

	lwi.items[level] = append(lwi.items[level], item)
}

// Set specifies the Items for an entire level.
func (lwi *LevelsWithItems) Set(level int, items []Item) {
	if lwi.items == nil {
		lwi.items = make(map[int][]Item)
	}

	lwi.items[level] = items
}

func (lwi *LevelsWithItems) itemsAt(level int) []Item {
	return lwi.items[level]
}

func (lwi *LevelsWithItems) removeItemsAt(level int) []Item {
	items := lwi.items[level]
	delete(lwi.items, level)
	return items
}
