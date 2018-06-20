package covertree

import "sync"

type InMemoryStore struct {
	items map[string]map[int][]Item
	mutex sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		items: make(map[string]map[int][]Item),
	}
}

func (s *InMemoryStore) Load(parent Item, level int) (items []Item, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.items[parent.CoverTreeID()][level], nil
}

func (s *InMemoryStore) Save(item, parent Item, level int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.levelsFor(item)

	id := item.CoverTreeID()
	levels := s.levelsFor(parent)
	for i, levelItem := range levels[level] {
		if levelItem.CoverTreeID() == id {
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
