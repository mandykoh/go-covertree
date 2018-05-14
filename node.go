package covertree

import (
	"math"
)

type Node struct {
	Item     Coverable
	children map[int][]*Node
}

func (n *Node) Children(level int) []*Node {
	children := append([]*Node{}, n)
	children = append(children, n.children[level]...)
	return children
}

func Insert(node *Node, coverSet []*Node, level int) bool {
	distThreshold := math.Pow(2, float64(level))

	var coverSetChildren []*Node
	for _, cn := range coverSet {
		for _, child := range cn.Children(level - 1) {
			if node.Item.Distance(child.Item) <= distThreshold {
				coverSetChildren = append(coverSetChildren, child)
			}
		}
	}
	if len(coverSetChildren) == 0 {
		return false
	}

	if Insert(node, coverSetChildren, level-1) {
		return true
	}

	for _, c := range coverSet {
		if node.Item.Distance(c.Item) <= distThreshold {
			if c.children == nil {
				c.children = make(map[int][]*Node)
			}
			c.children[level-1] = append(c.children[level-1], node)
			return true
		}
	}

	return false
}
