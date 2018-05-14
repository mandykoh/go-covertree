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

func Insert(node *Node, coverSet []*Node, level int) bool {
	distThreshold := math.Pow(2, float64(level))

	var coverSetChildren []*Node
	for _, cn := range coverSet {
		if node.Item.Distance(cn.Item) <= distThreshold {
			coverSetChildren = append(coverSetChildren, cn)
		}
		for _, child := range cn.children[level-1] {
			if node.Item.Distance(child.Item) <= distThreshold {
				coverSetChildren = append(coverSetChildren, child)
			}
		}
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
