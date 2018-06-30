package covertree

// ItemWithDistance represents an Item and its distance from some other
// predetermined item, as defined by a DistanceFunc.
type ItemWithDistance struct {
	Item     Item
	Distance float64
}
