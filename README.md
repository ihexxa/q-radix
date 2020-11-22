# q-radix

[![Build Status](https://travis-ci.org/ihexxa/q-radix.svg?branch=master)](https://travis-ci.org/ihexxa/q-radix)
[![Go Report](https://goreportcard.com/badge/github.com/ihexxa/q-radix)](https://goreportcard.com/report/github.com/ihexxa/q-radix)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ihexxa/q-radix)](https://pkg.go.dev/github.com/ihexxa/q-radix)

_A fast and simple (200+ lines) radix tree implementation in Go/Golang._

### Features

* good performance ([benchmark](https://github.com/ihexxa/radix-bench))
* more ways to query (Get, GetAllMatches, GetLongestMatch, BFS)
* simple interfaces
* well tested (unit tests and random tests)

### Examples

Document is [here](https://pkg.go.dev/github.com/ihexxa/q-radix).

```go
// import q-radix
import qradix "github.com/ihexxa/q-radix"


rTree := qradix.NewRTree() // create a new radix tree
ok := rTree.Insert("key", value) // insert value in any type with a string key
treeSize := rTree.Size() // get the size of the radix tree

val, ok := rTree.Get("key") // get the value by key
val, ok := rTree.GetAllMatches("key") // get all prefix matches of the key
val, ok := rTree.GetLongestMatch("key") // get the longest prefix match of the key

// override the value of the key if it exists in the radix tree
// and the old value will be returned if it exists in the radix tree
oldVal, ok := rTree.Insert("key", newValue)
ok := rTree.Remove("key") // remove the value from the radix tree

```
