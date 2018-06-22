package covertree

type coverSetItem struct {
	item     Item
	distance float64
}

func coverSetItemForQuery(item Item, query Item) coverSetItem {
	return coverSetItem{
		item:     item,
		distance: query.Distance(item),
	}
}
