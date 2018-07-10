package covertree

type inMemoryStore struct {
	distanceBetween DistanceFunc
	items           map[Item]map[int][]Item
}

func newInMemoryStore(distanceFunc DistanceFunc) *inMemoryStore {
	return &inMemoryStore{
		distanceBetween: distanceFunc,
		items:           make(map[Item]map[int][]Item),
	}
}

func (s *inMemoryStore) AddItem(item, parent Item, level int) error {
	s.levelsFor(item)
	levels := s.levelsFor(parent)
	levels[level] = append(levels[level], item)
	return nil
}

func (s *inMemoryStore) LoadChildren(parent Item) (result LevelsWithItems, err error) {
	for level, items := range s.items[parent] {
		result.Set(level, items)
	}

	return
}

func (s *inMemoryStore) RemoveItem(item, parent Item, level int) error {
	levels := s.items[parent]
	for i, levelItem := range levels[level] {
		if levelItem == item || s.distanceBetween(levelItem, item) == 0 {
			levels[level] = append(levels[level][:i], levels[level][i+1:]...)
			delete(s.items, item)
			return nil
		}
	}

	return nil
}

func (s *inMemoryStore) UpdateItem(item, parent Item, level int) error {
	if parent == nil {
		delete(s.items, parent)
	}

	levels := s.levelsFor(parent)
	levels[level] = append(levels[level], item)
	return nil
}

func (s *inMemoryStore) levelsFor(item Item) map[int][]Item {
	levels, ok := s.items[item]
	if !ok {
		levels = make(map[int][]Item)
		s.items[item] = levels
	}

	return levels
}

// NewInMemoryTree creates a new, empty tree which is backed by an in-memory
// store. The tree will use the specified function for determining the distance
// between items.
func NewInMemoryTree(distanceFunc DistanceFunc) *Tree {
	tree, _ := NewTreeWithStore(newInMemoryStore(distanceFunc), distanceFunc)
	return tree
}
