# qradix.go

*A simple radix tree implementation in Go.*

### Features

- Simple API, start quickly
- Fully tested

### Quick Start

```go
	// create a radix tree
	rTree := qradix.NewRTree()

	// insert value in any type with a string key
	ok := rTree.Insert("key", value)

	// get the value of key 
	val, ok := rTree.Get("key")

	// override the value of key
	rTree.Insert("key", newValue)

	// remove the value of key
	ok := rTree.Remove("key")

	// get the size of radix tree
	treeSize := rTree.Size
```