# q-radix

_A cute(Q) radix tree implementation in Go/Golang with good performance._

[![](https://travis-ci.org/ihexxa/q-radix.svg?branch=master)](build)
[![](https://goreportcard.com/badge/github.com/ihexxa/q-radix)](goreport)

### Features

* good performance
* more ways to query: GetAllMatches
* simple interfaces
* well tested (unit test and random test)

### Install

Install q-radix: `go get github.com/ihexxa/q-radix`

### Examples

```go
// import q-radix
import qradix "github.com/ihexxa/q-radix"


rTree := qradix.NewRTree() // create a new radix tree
ok := rTree.Insert("key", value) // insert value in any type with a string key
treeSize := rTree.Size // get the size of radix tree

val, ok := rTree.Get("key") // get the value by key
val, ok := rTree.GetAllMatches("key") // get all prefix matches of the key

// override the value of the key if it exists in the radix tree
// and the old value will be returned if it exists in the radix tree
oldVal, ok := rTree.Insert("key", newValue)
ok := rTree.Remove("key") // remove the value from the radix tree

```
