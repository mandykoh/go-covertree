package covertree

// LevelsWithItems represents a set of child Items of a parent Item, separated
// into their levels.
type LevelsWithItems struct {
	items map[int][]Item
}

// Add adds an Item to the specified level.
func (iwl *LevelsWithItems) Add(level int, item Item) {
	if iwl.items == nil {
		iwl.items = make(map[int][]Item)
	}

	iwl.items[level] = append(iwl.items[level], item)
}

// Set specifies the Items for an entire level.
func (iwl *LevelsWithItems) Set(level int, items []Item) {
	if iwl.items == nil {
		iwl.items = make(map[int][]Item)
	}

	iwl.items[level] = items
}

func (iwl *LevelsWithItems) itemsAt(level int) []Item {
	return iwl.items[level]
}

func (iwl *LevelsWithItems) removeItemsAt(level int) []Item {
	items := iwl.items[level]
	delete(iwl.items, level)
	return items
}
