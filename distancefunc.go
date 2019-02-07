package covertree

// DistanceFunc represents a function which defines the distance between two
// items.
type DistanceFunc func(a, b interface{}) (distance float64)
