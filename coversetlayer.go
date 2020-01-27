package covertree

import (
	"sort"
)

type coverSetLayer []itemWithChildren

func (l coverSetLayer) constrainedToDistance(distance float64) coverSetLayer {
	cutOff := sort.Search(len(l), func(i int) bool {
		return l[i].withDistance.Distance > distance
	})

	return l[:cutOff]
}

func makeCoverSetLayer(items []itemWithChildren) coverSetLayer {
	sort.Slice(items, func(i, j int) bool {
		return items[i].withDistance.Distance < items[j].withDistance.Distance
	})
	return items
}
