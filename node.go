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

func Insert(node *Node, coverSet coverSet, level int) bool {
	distThreshold := math.Pow(2, float64(level))
	childCoverSet := coverSet.child(node.Item, distThreshold, level-1)

	if len(childCoverSet) > 0 {
		if Insert(node, childCoverSet, level-1) {
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
