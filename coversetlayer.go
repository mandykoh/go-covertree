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
	layer := make(coverSetLayer, len(items))
	copy(layer, items)

	sort.Slice(layer, func(i, j int) bool {
		return layer[i].withDistance.Distance < layer[j].withDistance.Distance
	})

	return layer
}
