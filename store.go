package covertree

// Store implementations allow entire Trees to be made accessible in an
// extensible way. Implementations may provide capabilities such as persistence
// and serialisation to various formats and data stores.
//
// Each Store instance is responsible for storing and retrieval of the data for
// a single Tree. This implies that any keys or identifiers required for the
// loading and saving of a tree should be known to the instance of the store.
//
// Store implementations will typically need to know about the type of Items
// being stored in the tree.
type Store interface {

	// DeleteChild removes an item from the store as a child of the specified
	// parent item, at the given level. If no such item exists, this operation
	// should have no effect.
	//
	// Implementations are free to completely delete the item itself, but should
	// bear in mind that any children of the item should continue to exist and
	// will be re-parented to other items.
	DeleteChild(item, parent Item, level int) error

	// LoadChildren returns the explicit child items of the specified parent
	// item along with their levels.
	LoadChildren(parent Item) (LevelsWithItems, error)

	// LoadTree is called by NewTreeFromStore to retrieve the metadata for the
	// Tree instance being managed by the Store.
	LoadTree() (root Item, rootLevel int, err error)

	// SaveChild saves an item to the store as a child of the specified parent
	// item, at the given level.
	SaveChild(child, parent Item, level int) error

	// SaveTree is called by a Tree to save its metadata.
	SaveTree(root Item, rootLevel int) error
}
