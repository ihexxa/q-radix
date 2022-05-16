package main

import (
	"fmt"

	qradix "github.com/ihexxa/q-radix"
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
}
