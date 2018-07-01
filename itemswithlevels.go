package covertree

type ItemsWithLevels struct {
	items map[int][]Item
}

func (iwl *ItemsWithLevels) Add(level int, item Item) {
	if iwl.items == nil {
		iwl.items = make(map[int][]Item)
	}

	iwl.items[level] = append(iwl.items[level], item)
}

func (iwl *ItemsWithLevels) Set(level int, items []Item) {
	if iwl.items == nil {
		iwl.items = make(map[int][]Item)
	}

	iwl.items[level] = items
}

func (iwl *ItemsWithLevels) deleteLevel(level int) {
	delete(iwl.items, level)
}

func (iwl *ItemsWithLevels) itemsForLevel(level int) []Item {
	return iwl.items[level]
}
