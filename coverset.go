package covertree

type coverSet []*Node

func (cs *coverSet) child(item Coverable, distThreshold float64, childLevel int) (child coverSet) {
	for _, node := range *cs {
		if item.Distance(node.Item) <= distThreshold {
			child = append(child, node)
		}

		for _, childNode := range node.children[childLevel] {
			if item.Distance(childNode.Item) <= distThreshold {
				child = append(child, childNode)
			}
		}
	}

	return
}
