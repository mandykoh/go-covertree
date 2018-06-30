package covertree

type inMemoryStore struct {
	items map[Item]map[int][]Item
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{
		items: make(map[Item]map[int][]Item),
	}
}

func (s *inMemoryStore) LoadChildren(parent Item, level int) (children []Item, err error) {
	return s.items[parent][level], nil
}

func (s *inMemoryStore) LoadTree() (root Item, rootLevel, deepestLevel int, err error) {
	return nil, 0, 0, nil
}

func (s *inMemoryStore) SaveChild(child, parent Item, level int) error {
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

func (s *inMemoryStore) SaveTree(root Item, rootLevel, deepestLevel int) error {
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
	tree, _ := NewTreeFromStore(newInMemoryStore(), distanceFunc)
	return tree
}
