package covertree

type InMemoryStore struct {
	items map[string]map[int][]Item
}

func (s *InMemoryStore) Load(parent Item, level int) (items []Item, err error) {
	return s.items[parent.CoverTreeID()][level], nil
}

func (s *InMemoryStore) Save(item, parent Item, level int) error {
	if s.items == nil {
		s.items = make(map[string]map[int][]Item)
	}

	s.levelsFor(item)

	levels := s.levelsFor(parent)
	for i, levelItem := range levels[level] {
		if levelItem.CoverTreeID() == item.CoverTreeID() {
			levels[level][i] = item
			return nil
		}
	}

	levels[level] = append(levels[level], item)
	return nil
}

func (s *InMemoryStore) levelsFor(item Item) map[int][]Item {
	levels, ok := s.items[item.CoverTreeID()]
	if !ok {
		levels = make(map[int][]Item)
		s.items[item.CoverTreeID()] = levels
	}

	return levels
}
