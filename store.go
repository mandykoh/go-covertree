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
//
// Stores should implement distance-identity semantics; two items whose distance
// is exactly zero should be considered the same item.
type Store interface {

	// DeleteChild disassociates an item in the store from the specified parent
	// at the given level. If no such item exists, this operation should have no
	// effect.
	//
	// Implementations are free to completely delete the item itself along with
	// any relationships to child items, but should bear in mind that children
	// should continue to exist as orphans and will be re-parented to other
	// items.
	DeleteChild(item, parent Item, level int) error

	// LoadChildren returns the explicit child items of the specified parent
	// item along with their levels.
	LoadChildren(parent Item) (LevelsWithItems, error)

	// LoadTree is called by NewTreeFromStore to retrieve the metadata for the
	// Tree instance being managed by the Store.
	LoadTree() (root Item, rootLevel int, err error)

	// SaveChild saves an item to the store as a child of the specified parent
	// item, at the given level. It is valid for child items to be re-parented;
	// if the child already exists in the store, it becomes associated with the
	// new parent and level.
	//
	// Implementations are free to assume that this will only be called for new
	// or orphaned child items; cleanup of existing associations with parent
	// items should be performed by DeleteChild.
	SaveChild(child, parent Item, level int) error

	// SaveTree is called by a Tree to save its metadata.
	SaveTree(root Item, rootLevel int) error
}
