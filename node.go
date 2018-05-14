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
	var coverSetChildren []*Node
	for _, c := range coverSet {
		coverSetChildren = append(coverSetChildren, c.Children(level-1)...)
	}

	distThreshold := math.Pow(2, float64(level))

	minFound := false
	minDist := 0.0
	for _, c := range coverSetChildren {
		dist := node.Item.Distance(c.Item)
		if dist < minDist || !minFound {
			minDist = dist
			minFound = true
		}
	}
	if minFound && minDist > distThreshold {
		return false
	}

	var childCoverSet []*Node
	for _, c := range coverSetChildren {
		if node.Item.Distance(c.Item) <= distThreshold {
			childCoverSet = append(childCoverSet, c)
		}
	}

	if Insert(node, childCoverSet, level-1) {
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
