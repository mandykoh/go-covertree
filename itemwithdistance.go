package covertree

// ItemWithDistance represents an item and its distance from some other
// predetermined item, as defined by a DistanceFunc.
type ItemWithDistance struct {
	Item     interface{}
	Distance float64
}
