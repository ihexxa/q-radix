# q-radix

*A simple radix tree implementation in Go.*

### Features
"q" is a vision, which stands for good quality, quick and cute. 

- well tested (random tested, unit tested)
- good performance
- simple API

### Install
Install q-radix: `go get github.com/ihexxa/q-radix`


### Usages
```go
// import q-radix
import qradix "github.com/ihexxa/q-radix" 

// create a radix tree
rTree := qradix.NewRTree()

// insert value in any type with a string key
ok := rTree.Insert("key", value)

// get the value of key
// old value will be returned if key has already existed in radix tree
val, ok := rTree.Get("key")

// override the value of key
rTree.Insert("key", newValue)

// remove the value of key
ok := rTree.Remove("key")

// get the size of radix tree
treeSize := rTree.Size
```
