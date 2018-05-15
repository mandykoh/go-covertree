package covertree

import (
	"math"
)

type Node struct {
	Item     Coverable
	Children map[int][]Node
}

func (n *Node) addChild(item Coverable, level int) {
	if n.Children == nil {
		n.Children = make(map[int][]Node)
	}
	n.Children[level] = append(n.Children[level], Node{Item: item})
}

func Insert(item Coverable, coverSet coverSet, level int) bool {
	distThreshold := math.Pow(2, float64(level))
	childCoverSet := coverSet.child(item, distThreshold, level-1)

	if len(childCoverSet) > 0 {
		if Insert(item, childCoverSet, level-1) {
			return true
		}

		for _, cn := range coverSet {
			if item.Distance(cn.Item) <= distThreshold {
				cn.addChild(item, level-1)
				return true
			}
		}
	}

	return false
}
