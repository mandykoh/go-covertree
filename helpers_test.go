package covertree

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func compareWithLinearSearch(tree *Tree, points []Point, maxResults int, maxDistance float64, coverTreeDistanceCalls *int, t *testing.T) {
	t.Helper()

	query := randomPoint()
	fmt.Printf("Query point %v (maxResults: %d, maxDistance: %g)\n", query, maxResults, maxDistance)

	*coverTreeDistanceCalls = 0

	startTime := time.Now()
	coverTreeResults, err := tree.FindNearest(&query, maxResults, maxDistance)
	finishTime := time.Now()

	if err != nil {
		t.Fatalf("Error querying tree: %v", err)
	}

	fmt.Printf("Cover Tree FindNearest took %d distance comparisons, %dms\n", *coverTreeDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)
	for _, r := range coverTreeResults {
		point := *(r.Item.(*Point))
		fmt.Printf("Cover Tree FindNearest: %v (distance %g)\n", point, r.Distance)
	}

	linearSearchResults, linearSearchDistanceCalls := linearSearch(&query, points, maxResults, maxDistance)

	expectSameResults(t, query, coverTreeResults, linearSearchResults)

	if *coverTreeDistanceCalls >= linearSearchDistanceCalls {
		t.Errorf("Expected cover tree search to require fewer than %d distance comparisons (linear search) but got %d", linearSearchDistanceCalls, coverTreeDistanceCalls)
	}

	return
}

func distanceBetweenPoints(a, b interface{}) float64 {
	p1 := a.(*Point)
	p2 := b.(*Point)

	total := 0.0
	for i := 0; i < len(p1); i++ {
		diff := p2[i] - p1[i]
		total += diff * diff
	}

	return math.Sqrt(total)
}

func distanceBetweenPointsWithCounter(counter *int) DistanceFunc {
	return func(a, b interface{}) float64 {
		*counter++
		return distanceBetweenPoints(a, b)
	}
}

func expectSameResults(t *testing.T, query Point, actualResults []ItemWithDistance, expectedResults []ItemWithDistance) {
	t.Helper()

	if expected, actual := len(expectedResults), len(actualResults); expected != actual {
		t.Fatalf("Expected %d results for %v but got %d instead", expected, query, actual)

	}

	availableResults := len(expectedResults)
	if len(actualResults) < availableResults {
		availableResults = len(actualResults)
	}

	for i := 0; i < availableResults; i++ {
		expectedResult := expectedResults[i]
		actualResult := actualResults[i]

		if expected, actual := expectedResult.Item, actualResult.Item; expected != actual {
			t.Errorf("Expected nearest point %d to %v to be %v but got %v", i, query, *expected.(*Point), *actual.(*Point))
		}
		if expected, actual := expectedResult.Distance, actualResult.Distance; expected != actual {
			t.Errorf("Expected distance of nearest point %d to %v to be %v but got %v", i, query, expected, actual)
		}
	}
}

func insertPoints(points []Point, tree *Tree) (timeTaken time.Duration, err error) {
	startTime := time.Now()

	for i := range points {
		err := tree.Insert(&points[i])
		if err != nil {
			return 0, err
		}
	}

	finishTime := time.Now()

	return finishTime.Sub(startTime), nil
}

func linearSearch(query *Point, points []Point, maxResults int, maxDistance float64) (results []ItemWithDistance, distanceCallCount int) {
	distanceCalls := 0
	distanceBetween := distanceBetweenPointsWithCounter(&distanceCalls)

	startTime := time.Now()

	results = make([]ItemWithDistance, maxResults, maxResults)

	for i := range points {
		dist := distanceBetween(query, &points[i])
		if dist > maxDistance {
			continue
		}

		for j := 0; j < len(results); j++ {
			if results[j].Item == nil || dist < results[j].Distance {
				for k := len(results) - 1; k > j; k-- {
					results[k] = results[k-1]
				}
				results[j].Item = &points[i]
				results[j].Distance = dist
				break
			}
		}
	}

	lastNonNil := len(results) - 1
	for lastNonNil >= 0 && results[lastNonNil].Item == nil {
		lastNonNil--
	}
	results = results[:lastNonNil+1]

	finishTime := time.Now()

	linearSearchDistanceCalls := distanceCalls

	fmt.Printf("Linear FindNearest took %d distance comparisons, %dms\n", linearSearchDistanceCalls, finishTime.Sub(startTime)/time.Millisecond)

	for _, r := range results {
		fmt.Printf("Linear FindNearest: %v (distance %g)\n", *r.Item.(*Point), r.Distance)
	}

	return results, linearSearchDistanceCalls
}

func randomPoint() (point Point) {
	for i := 0; i < len(point); i++ {
		point[i] = rand.Float64() * 1000
	}
	return point
}

func randomPoints(count int) (points []Point) {
	pointsMap := make(map[Point]bool, count)
	for len(pointsMap) < count {
		val := randomPoint()
		if _, exists := pointsMap[val]; !exists {
			pointsMap[val] = true
			points = append(points, val)
		}
	}

	return
}

func traverseNodes(item, parent interface{}, level int, indentLevel int, store *inMemoryStore, print bool) (nodeCount int) {
	if print {
		fmt.Printf("%4d: ", level)
		for i := 0; i < indentLevel; i++ {
			fmt.Print("..")
		}
		if indentLevel > 0 {
			fmt.Print(" ")
		}

		fmt.Printf("%v", item)

		if parent != nil {
			fmt.Printf(" (%g)", distanceBetweenPoints(item, parent))
		}

		fmt.Println()
	}

	nodeCount = 1

	var levels []int
	for k := range store.levelsFor(item) {
		levels = append(levels, k)
	}
	sort.Ints(levels)

	for i := len(levels) - 1; i >= 0; i-- {
		l := levels[i]
		children, _ := store.LoadChildren(item)
		for _, c := range children.itemsAt(l) {
			nodeCount += traverseNodes(c, item, l, indentLevel+1, store, print)
		}
	}

	return
}

func traverseTree(tree *Tree, store *inMemoryStore, print bool) (nodeCount int) {
	if print {
		fmt.Println("---")
	}

	root, rootLevel, _ := tree.loadRoot()

	if root == nil {
		return 0
	}

	return traverseNodes(root, nil, rootLevel, 0, store, print)
}

type testStore struct {
	inMemoryStore
	savedCount     int
	savedRoot      interface{}
	savedRootLevel int
}

func newTestStore(distanceFunc DistanceFunc) *testStore {
	return &testStore{inMemoryStore: *newInMemoryStore(distanceFunc)}
}

func (ts *testStore) AddItem(item, parent interface{}, level int) error {
	if parent == nil {
		ts.savedCount++
		ts.savedRoot = item
		ts.savedRootLevel = level
	}
	return ts.inMemoryStore.AddItem(item, parent, level)
}

func (ts *testStore) RemoveItem(item, parent interface{}, level int) error {
	if parent == nil {
		ts.savedCount++
		ts.savedRoot = nil
	}
	return ts.inMemoryStore.RemoveItem(item, parent, level)
}

func (ts *testStore) UpdateItem(item, parent interface{}, level int) error {
	if parent == nil {
		ts.savedCount++
		ts.savedRoot = item
		ts.savedRootLevel = level
	}
	return ts.inMemoryStore.UpdateItem(item, parent, level)
}

func (ts *testStore) expectSavedTree(t *testing.T, saveCount int, root interface{}, rootLevel int) {
	t.Helper()

	if expected, actual := saveCount, ts.savedCount; expected != actual {
		t.Errorf("Expected tree to have been saved %d times but was saved %d times instead", expected, actual)
	}
	if expected, actual := root, ts.savedRoot; expected != actual {
		t.Errorf("Expected tree root to be %v but was %v", expected, actual)
	}
	if expected, actual := rootLevel, ts.savedRootLevel; expected != actual {
		t.Errorf("Expected tree root level to be at %d but was %d", expected, actual)
	}
}

type Point [3]float64

func (p *Point) String() string {
	return fmt.Sprintf("[%g %g %g]", p[0], p[1], p[2])
}
