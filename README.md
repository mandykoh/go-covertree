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

Define a [`DistanceFunc`](distancefunc.go) to compute the distance between two instances of the type:

```go
func distanceBetween(a, b Item) float64 {
    p1 := a.(*Point)
    p2 := b.(*Point)
	
    distX := math.Abs(p1.X - p2.X)
    distY := math.Abs(p1.Y - p2.Y)
	
    return math.Sqrt(distX * distX + distY * distY)
}
```



