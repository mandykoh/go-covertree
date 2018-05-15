package covertree

type coverSet []*Node

func (cs *coverSet) child(item Item, distThreshold float64, childLevel int) (child coverSet) {
	for _, node := range *cs {

		if item.Distance(node.Item) <= distThreshold {
			child = append(child, node)
		}

		children := node.Children[childLevel]

		for i := 0; i < len(children); i++ {
			childNode := &children[i]
			if item.Distance(childNode.Item) <= distThreshold {
				child = append(child, childNode)
			}
		}
	}

	return
}
