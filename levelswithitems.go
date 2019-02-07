package covertree

// LevelsWithItems represents a set of child items of a parent item, separated
// into their levels.
type LevelsWithItems struct {
	items map[int][]interface{}
}

// Add adds an item to the specified level.
func (lwi *LevelsWithItems) Add(level int, item interface{}) {
	if lwi.items == nil {
		lwi.items = make(map[int][]interface{})
	}

	lwi.items[level] = append(lwi.items[level], item)
}

// Set specifies the items for an entire level.
func (lwi *LevelsWithItems) Set(level int, items []interface{}) {
	if lwi.items == nil {
		lwi.items = make(map[int][]interface{})
	}

	lwi.items[level] = items
}

func (lwi *LevelsWithItems) itemsAt(level int) []interface{} {
	return lwi.items[level]
}

func (lwi *LevelsWithItems) removeItemsAt(level int) []interface{} {
	items := lwi.items[level]
	delete(lwi.items, level)
	return items
}
