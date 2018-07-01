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

	// LoadChildren returns the explicit child items of the specified parent
	// item at a given level.
	LoadChildren(parent Item, level int) (children []Item, err error)

	// LoadTree is called by NewTreeFromStore to retrieve the metadata for the
	// Tree instance being managed by the Store.
	LoadTree() (root Item, rootLevel, deepestLevel int, err error)

	// SaveChild saves an item to the store as a child of the specified parent
	// item, at the given level.
	SaveChild(child, parent Item, level int) error

	// SaveTree is called by a Tree to save its metadata.
	SaveTree(root Item, rootLevel, deepestLevel int) error
}
