package covertree

import "sync"

type inMemoryStore struct {
	distanceBetween DistanceFunc
	items           map[interface{}]map[int][]interface{}
	mutex           sync.RWMutex
}

func newInMemoryStore(distanceFunc DistanceFunc) *inMemoryStore {
	return &inMemoryStore{
		distanceBetween: distanceFunc,
		items:           make(map[interface{}]map[int][]interface{}),
	}
}

func (s *inMemoryStore) AddItem(item, parent interface{}, level int) error {
	return s.UpdateItem(item, parent, level)
}

func (s *inMemoryStore) LoadChildren(parent interface{}) (result LevelsWithItems, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for level, items := range s.items[parent] {
		result.Set(level, items)
	}

	return
}

func (s *inMemoryStore) RemoveItem(item, parent interface{}, level int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	levels := s.items[parent]
	for i, levelItem := range levels[level] {
		if levelItem == item {
			levels[level] = append(levels[level][:i], levels[level][i+1:]...)
			delete(s.items, item)
			return nil
		}
	}

	return nil
}

func (s *inMemoryStore) UpdateItem(item, parent interface{}, level int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if parent == nil {
		delete(s.items, parent)
	}

	levels := s.levelsFor(parent)
	levels[level] = append(levels[level], item)
	return nil
}

func (s *inMemoryStore) levelsFor(item interface{}) map[int][]interface{} {
	levels, ok := s.items[item]
	if !ok {
		levels = make(map[int][]interface{})
		s.items[item] = levels
	}

	return levels
}

// NewInMemoryTree creates a new, empty tree which is backed by an in-memory
// store. The tree will use the specified function for determining the distance
// between items.
//
// Note that for the sake of efficiency, and due to how an in-memory tree will
// tend to be used, the in-memory implementation uses pointer equality instead
// of distance-identity. In particular, this means that removal requires the
// exact item to be specified in order to be removed.
func NewInMemoryTree(basis float64, distanceFunc DistanceFunc) *Tree {
	tree, _ := NewTreeWithStore(newInMemoryStore(distanceFunc), basis, distanceFunc)
	return tree
}
