package covertree

type InMemoryStore struct {
	items map[Item]map[int][]Item
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items: make(map[Item]map[int][]Item),
	}
}

func (s *InMemoryStore) Load(parent Item, level int) (items []Item, err error) {
	return s.items[parent][level], nil
}

func (s *InMemoryStore) Save(item, parent Item, level int) error {
	s.levelsFor(item)

	levels := s.levelsFor(parent)
	for i, levelItem := range levels[level] {
		if levelItem == item {
			levels[level][i] = item
			return nil
		}
	}

	levels[level] = append(levels[level], item)
	return nil
}

func (s *InMemoryStore) levelsFor(item Item) map[int][]Item {
	levels, ok := s.items[item]
	if !ok {
		levels = make(map[int][]Item)
		s.items[item] = levels
	}

	return levels
}
