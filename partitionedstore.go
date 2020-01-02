package covertree

import (
	"hash/fnv"
	"sync"
)

type PartitioningFunc func(parentItem interface{}) (partitionKey string)

type partitionedStore struct {
	partitionForParent PartitioningFunc
	stores             []Store
}

func (s *partitionedStore) AddItem(item, parent interface{}, level int) error {
	store, err := s.storeForParent(parent)
	if err != nil {
		return err
	}

	return store.AddItem(item, parent, level)
}

func (s *partitionedStore) LoadChildren(parents ...interface{}) (children []LevelsWithItems, err error) {

	entriesByStore := make(map[Store]struct {
		parents       []interface{}
		parentIndices []int
	})

	for i, parent := range parents {
		store, err := s.storeForParent(parent)
		if err != nil {
			return nil, err
		}

		entry := entriesByStore[store]
		entry.parents = append(entry.parents, parent)
		entry.parentIndices = append(entry.parentIndices, i)
		entriesByStore[store] = entry
	}

	children = make([]LevelsWithItems, len(parents))

	var doneGroup sync.WaitGroup
	doneGroup.Add(len(entriesByStore))

	var nestedErr error
	for storeForEntry, entryForStore := range entriesByStore {
		store := storeForEntry
		entry := entryForStore

		go func() {
			defer doneGroup.Done()

			c, err := store.LoadChildren(entry.parents...)
			if err != nil {
				nestedErr = err
				return
			}

			for i := range c {
				children[entry.parentIndices[i]] = c[i]
			}
		}()
	}

	doneGroup.Wait()

	return children, nestedErr
}

func (s *partitionedStore) RemoveItem(item, parent interface{}, level int) error {
	store, err := s.storeForParent(parent)
	if err != nil {
		return err
	}

	return store.RemoveItem(item, parent, level)
}

func (s *partitionedStore) UpdateItem(item, parent interface{}, level int) error {
	store, err := s.storeForParent(parent)
	if err != nil {
		return err
	}

	return store.UpdateItem(item, parent, level)
}

func (s *partitionedStore) storeForParent(parent interface{}) (Store, error) {
	partitionKey := s.partitionForParent(parent)

	hash := fnv.New32a()
	_, err := hash.Write([]byte(partitionKey))
	if err != nil {
		return nil, err
	}

	return s.stores[hash.Sum32()%uint32(len(s.stores))], nil
}

// NewPartitionedStore returns a store which distributes store operations across
// the underlying stores using the specified partitioning function.
//
// Operations for a given partition key are always assigned to the same store.
func NewPartitionedStore(partitioningFunc PartitioningFunc, stores ...Store) *partitionedStore {
	return &partitionedStore{
		partitionForParent: partitioningFunc,
		stores:             stores,
	}
}
