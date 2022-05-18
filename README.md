# q-radix

[![Build Status](https://travis-ci.org/ihexxa/q-radix.svg?branch=master)](https://travis-ci.org/ihexxa/q-radix)
[![Go Report](https://goreportcard.com/badge/github.com/ihexxa/q-radix)](https://goreportcard.com/report/github.com/ihexxa/q-radix)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ihexxa/q-radix)](https://pkg.go.dev/github.com/ihexxa/q-radix)

_A simple and serializable radix tree implementation in Go/Golang._

### Features

- Simple APIs: Insert, Get, Remove, GetAllPrefixMatches, GetBestMatch.
- Serializable: String() method is supported, then it can be persisted.
- UTF-8 support: support different characters as keys
- Well tested: it is covered by unit tests and random tests.
- Good performance: [benchmark](https://github.com/ihexxa/radix-bench).

### Examples

Document is [here](https://pkg.go.dev/github.com/ihexxa/q-radix).

```go
package main

import (
	"fmt"

	qradix "github.com/ihexxa/q-radix/v3"
)

func main() {
	// create a new radix tree
	rTree := qradix.NewRTree()

	// insert value in any type with a string key
	_, err := rTree.Insert("key", "value")
	if err != nil {
		panic(err)
	}

	// get the value by key
	val, err := rTree.Get("key")
	if err != nil {
		panic(err)
	}

	// get all prefix matches of the key
	keyValues := rTree.GetAllPrefixMatches("key")
	if err != nil {
		panic(err)
	}
	for key, value := range keyValues {
		fmt.Println(key, value)
	}

	// get the longest prefix match of the key
	key, val, ok := rTree.GetBestMatch("key")
	fmt.Println(key, val, ok)

	// override the value of the key if it exists in the radix tree
	// and the old value will be returned if it exists
	oldValue, err := rTree.Insert("key", "newValue")
	ok = rTree.Remove("key") // remove the value from the radix tree
	fmt.Println(oldValue, err, ok)

	// get the size of the radix tree
	fmt.Println(rTree.Size())

	// traverse the tree
	rTree.Insert("he", "v1")
	rTree.Insert("hello", "v2")
	rTree.Insert("hello世界", "v3")
	qradix.BFS(rTree, qradix.PrintNode)

	// serialize tree into rows!
	rows := []string{}
	for row := range rTree.String() {
		rows = append(rows, row)
	}

	// restore a new tree from rows!
	rTree2 := qradix.NewRTree()
	rowChan := make(chan string, len(rows))
	go func() {
		for _, row := range rows {
			rowChan <- row
		}
		close(rowChan)
	}()
	rTree2.FromString(rowChan)
	qradix.BFS(rTree2, qradix.PrintNode)
}
```
