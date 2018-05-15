package covertree

import (
	"fmt"
	"math"
	"sort"
	"testing"
)

type Point struct {
	x float64
}

func (p Point) CoverTreeID() string {
	return fmt.Sprintf("%0.0f", p.x)
}

func (p Point) Distance(other Item) float64 {
	op := other.(Point)
	return math.Abs(op.x - p.x)
}

func PrintTree(n *Node, level int, indentLevel int) {
	fmt.Printf("%4d: ", level)
	for i := 0; i < indentLevel; i++ {
		fmt.Print("  ")
	}

	fmt.Println(n.Item)

	var levels []int
	for k := range n.Children {
		levels = append(levels, k)
	}
	sort.Ints(levels)

	for i := len(levels) - 1; i >= 0; i-- {
		l := levels[i]
		children := n.Children[l]
		for i := range children {
			PrintTree(&children[i], l, indentLevel+1)
		}
	}
}

func TestSomething(t *testing.T) {

	root := Node{
		Item: Point{1},
	}

	for i := 1; i < 20; i++ {
		val := float64(i)/10.0 + 1
		result := Insert(Point{val}, coverSet{&root}, 10)
		fmt.Println("Result", result)
	}

	PrintTree(&root, 10, 0)
}
