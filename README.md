# go-covertree

[![GoDoc](https://godoc.org/github.com/mandykoh/go-covertree?status.svg)](https://godoc.org/github.com/mandykoh/go-covertree)
[![Go Report Card](https://goreportcard.com/badge/github.com/mandykoh/go-covertree)](https://goreportcard.com/report/github.com/mandykoh/go-covertree)
[![Build Status](https://travis-ci.org/mandykoh/go-covertree.svg?branch=master)](https://travis-ci.org/mandykoh/go-covertree)

`go-covertree` is a [cover tree](http://hunch.net/~jl/projects/cover_tree/icml_final/final-icml.pdf) implementation in Go for nearest-neighbour search and clustering. It uses an extensible backing store interface (suitable to adapting to key-value stores, RDBMSes, etc) to support very large data sets.

See the [API documentation](https://godoc.org/github.com/mandykoh/go-covertree) for more details.

This software is made available under an [MIT license](LICENSE).


## Thread safety

[Tree](https://godoc.org/github.com/mandykoh/go-covertree#Tree) instances are thread-safe for readonly access.

Insertions into the tree (using `Insert`) are purely append-only operations, and safe to make concurrently, allowing tree construction to be parallelised.

Searching the tree (using `FindNearest`) is purely a read-only operation and safe to do concurrently, including with insertions.

Removals from the tree (using `Remove`) are not thread-safe and should be externally synchronised if concurrent read-write access is required.

[Store](https://godoc.org/github.com/mandykoh/go-covertree#Store) implementations should observe their own thread-safety considerations.


## Example usage

Define a type to be stored in the tree:

```go
type Point struct {
    X float64
    Y float64
}
```

Define a [`DistanceFunc`](https://godoc.org/github.com/mandykoh/go-covertree#DistanceFunc) to compute the distance between two instances of the type:

```go
func distanceBetween(a, b interface{}) float64 {
    p1 := a.(*Point)
    p2 := b.(*Point)
	
    distX := p1.X - p2.X
    distY := p1.Y - p2.Y
	
    return math.Sqrt(distX * distX + distY * distY)
}
```

Create a [`Tree`](https://godoc.org/github.com/mandykoh/go-covertree#Tree). A tree using a provided in-memory store can be conveniently created using [`NewInMemoryTree`](https://godoc.org/github.com/mandykoh/go-covertree#NewInMemoryTree):

```go
tree := covertree.NewInMemoryTree(basis, rootDistance, distanceBetween)
```

The `basis` specifies the logarithmic base for determining the ratio of coverage of nodes at adjacent levels of the tree. If unsure, values around 2.0 may be good starting points.

The `rootDistance` specifies the maximum expected distance between nodes (actually the minimum distance between root nodes) and determines when new root nodes are created. This should generally be set to the largest distance between nodes expected for your data set.

Custom [`Store`](https://godoc.org/github.com/mandykoh/go-covertree#Store) implementations can also use the basic Tree constructor to create trees:

```go
// Creates a tree that is backed by a specific store
tree, err := covertree.NewTreeWithStore(pointStore, basis, rootDistance, distanceBetween)       
```

[Insert](https://godoc.org/github.com/mandykoh/go-covertree#Tree.Insert) some things into the tree:

```go
err := tree.Insert(&Point{1.5, 3.14})
```

[Find](https://godoc.org/github.com/mandykoh/go-covertree#Tree.FindNearest) the 5 nearest things in the store that are within 10.0 of a query point:

```go
results, err := tree.FindNearest(&Point{0.0, 0.0}, 5, 10.0)
```

[Remove](https://godoc.org/github.com/mandykoh/go-covertree#Tree.Remove) things from the store:

```go
removed, err := tree.Remove(&Point{1.5, 3.14})
```
