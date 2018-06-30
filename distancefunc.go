package covertree

// DistanceFunc represents a function which defines the distance between two
// Items.
type DistanceFunc func(a, b Item) (distance float64)
