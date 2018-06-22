package covertree

type Item interface {
	Distance(other Item) float64
}
