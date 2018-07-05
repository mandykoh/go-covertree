# go-covertree

[![GoDoc](https://godoc.org/github.com/mandykoh/go-covertree?status.svg)](https://godoc.org/github.com/mandykoh/go-covertree)
[![Go Report Card](https://goreportcard.com/badge/github.com/mandykoh/go-covertree)](https://goreportcard.com/report/github.com/mandykoh/go-covertree)
[![Build Status](https://travis-ci.org/mandykoh/go-covertree.svg?branch=master)](https://travis-ci.org/mandykoh/go-covertree)

`go-covertree` is a [cover tree](http://hunch.net/~jl/projects/cover_tree/icml_final/final-icml.pdf) implementation in Go for nearest-neighbour search and clustering. It uses an extensible backing store interface (suitable to adapting to key-value stores, RDBMSes, etc) to support very large data sets.

See the [API documentation](https://godoc.org/github.com/mandykoh/go-covertree) for more details.

This software is made available under an [MIT license](LICENSE).

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
func distanceBetween(a, b Item) float64 {
    p1 := a.(*Point)
    p2 := b.(*Point)
	
    distX := math.Abs(p1.X - p2.X)
    distY := math.Abs(p1.Y - p2.Y)
	
    return math.Sqrt(distX * distX + distY * distY)
}
```

Create a [`Tree`](https://godoc.org/github.com/mandykoh/go-covertree#Tree). A tree using a provided in-memory store can be conveniently created using [`NewInMemoryTree`](https://godoc.org/github.com/mandykoh/go-covertree#NewInMemoryTree):

```go
tree := covertree.NewInMemoryTree(distanceBetween)
```

Custom [`Store`](https://godoc.org/github.com/mandykoh/go-covertree#Store) implementations can also use the basic Tree constructors to create trees:

```go
tree, err := covertree.NewEmptyTreeWithStore(pointStore, distanceBetween)  // Creates a new empty tree
tree, err := covertree.NewTreeFromStore(pointStore, distanceBetween)       // Creates a tree that loads itself from a store
```

[Insert](https://godoc.org/github.com/mandykoh/go-covertree#Tree.Insert) some things into the tree:

```go
inserted, err := tree.Insert(&Point{1.5, 3.14})
```

[Find](https://godoc.org/github.com/mandykoh/go-covertree#Tree.FindNearest) the 5 nearest things in the store that are within 10.0 of a query point:

```go
results, err := tree.FindNearest(&Point{0.0, 0.0}, 5, 10.0)
```

[Remove](https://godoc.org/github.com/mandykoh/go-covertree#Tree.Remove) things from the store:

```go
err := tree.Remove(&Point{1.5, 3.14})
```
