package covertree

type coverSetItem struct {
	item           Item
	cachedDistance float64
}

func makeCoverSetItem(item Item) coverSetItem {
	return coverSetItem{
		item:           item,
		cachedDistance: -1,
	}
}

func (csi *coverSetItem) Distance(query Item) float64 {
	if csi.cachedDistance < 0 {
		csi.cachedDistance = csi.item.Distance(query)
	}
	return csi.cachedDistance
}
