package covertree

type ItemWithDistance struct {
	Item     Item
	Distance float64
}

func itemWithDistanceForQuery(item Item, query Item) ItemWithDistance {
	return ItemWithDistance{
		Item:     item,
		Distance: query.Distance(item),
	}
}
