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

	// AddItem saves an item to the store as a child of the specified parent
	// item, at the given level. The parent may be nil, indicating that an item
	// is being added at the root of the tree.
	//
	// Implementations are free to assume that this will only be called for new,
	// never-before-seen items.
	AddItem(item, parent Item, level int) error

	// LoadChildren returns the explicit child items of the specified parent
	// item along with their levels. If parent is nil, this is expected to
	// return the root item (which is a child of nil).
	LoadChildren(parent Item) (LevelsWithItems, error)

	// RemoveItem disassociates an item in the store from the specified parent
	// at the given level. If no such item exists, this operation should have no
	// effect.
	//
	// Implementations are free to completely delete the item itself along with
	// any relationships to child items, but should bear in mind that children
	// should continue to exist as orphans and will be re-parented to other
	// items.
	RemoveItem(item, parent Item, level int) error

	// UpdateItem updates the parent and level of a given item. It is valid for
	// child items to be re-parented; the item, which must already exist in the
	// store, is associated with the new parent and level.
	//
	// Implementations are free to assume that this will only be called for
	// items which have previously been added via AddItem.
	UpdateItem(item, parent Item, level int) error
}
