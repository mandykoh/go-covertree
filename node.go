package covertree

import (
	"math"
)

type Node struct {
	Item     Coverable
	children map[int][]*Node
}

func (n *Node) addChild(c *Node, level int) {
	if n.children == nil {
		n.children = make(map[int][]*Node)
	}
	n.children[level] = append(n.children[level], c)
}

func (n *Node) getChildrenWithinDistance(node *Node, distThreshold float64, childLevel int, dest *[]*Node) {
	if node.Item.Distance(n.Item) <= distThreshold {
		*dest = append(*dest, n)
	}

	for _, child := range n.children[childLevel] {
		if node.Item.Distance(child.Item) <= distThreshold {
			*dest = append(*dest, child)
		}
	}
}

func Insert(node *Node, coverSet []*Node, level int) bool {
	distThreshold := math.Pow(2, float64(level))

	var coverSetChildren []*Node
	for _, cn := range coverSet {
		cn.getChildrenWithinDistance(node, distThreshold, level-1, &coverSetChildren)
	}

	if len(coverSetChildren) > 0 {
		if Insert(node, coverSetChildren, level-1) {
			return true
		}

		for _, cn := range coverSet {
			if node.Item.Distance(cn.Item) <= distThreshold {
				cn.addChild(node, level-1)
				return true
			}
		}
	}

	return false
}
