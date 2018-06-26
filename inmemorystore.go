package covertree

type InMemoryStore struct {
	items map[Item]map[int][]Item
}

func newInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items: make(map[Item]map[int][]Item),
	}
}

func (s *InMemoryStore) LoadChildren(parent Item, level int) (children []Item, err error) {
	return s.items[parent][level], nil
}

func (s *InMemoryStore) SaveChild(child, parent Item, level int) error {
	s.levelsFor(child)

	levels := s.levelsFor(parent)
	for i, levelItem := range levels[level] {
		if levelItem == child {
			levels[level][i] = child
			return nil
		}
	}

	levels[level] = append(levels[level], child)
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

func NewInMemoryTree() *Tree {
	return NewTreeWithStore(newInMemoryStore())
}
