package covertree

type Item interface {
	CoverTreeID() string
	Distance(other Item) float64
}
