package covertree

type Coverable interface {
	Distance(item Coverable) float64
}
