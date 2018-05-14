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

func (p Point) Distance(other Coverable) float64 {
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
	for k := range n.children {
		levels = append(levels, k)
	}
	sort.Ints(levels)

	for i := len(levels) - 1; i >= 0; i-- {
		l := levels[i]
		for _, c := range n.children[l] {
			PrintTree(c, l, indentLevel+1)
		}
	}
}

func TestSomething(t *testing.T) {

	root := &Node{
		Item: Point{1},
	}

	result := Insert(&Node{Item: Point{1.1}}, []*Node{root}, 10)
	fmt.Println("Result", result)

	result = Insert(&Node{Item: Point{1.5}}, []*Node{root}, 10)
	fmt.Println("Result", result)

	result = Insert(&Node{Item: Point{1.05}}, []*Node{root}, 10)
	fmt.Println("Result", result)

	result = Insert(&Node{Item: Point{7.05}}, []*Node{root}, 10)
	fmt.Println("Result", result)

	result = Insert(&Node{Item: Point{1.52}}, []*Node{root}, 10)
	fmt.Println("Result", result)

	result = Insert(&Node{Item: Point{1.053}}, []*Node{root}, 10)
	fmt.Println("Result", result)

	PrintTree(root, 10, 0)
}
