package qradix

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const (
	insertAction = "insert"
	removeAction = "remove"
)

var (
	// use -args to input these options: "go test -args -d=true"
	actionCount = flag.Int("c", 10, "how many actions will be applied on RTree and map")
	insertRatio = flag.Int("i", 60, "control the the ratio between insert action and remove action")
	maxLen      = flag.Int("l", 5, "how long will random string be generated")
	seed        = flag.Int64("s", 10, "seed can be set to specific value to re-produce test failure")
	testRound   = flag.Int("r", 1, "how many times will random test run")
	treeHeight  = flag.Int("h", 3, "how many times will random suffix be append to root node's string")
	treeWidth   = flag.Int("w", 3, "how many random suffix will be append after node's string")
)

// TestWithRandomKeys starts a random test by comparing a RTree and a map
func TestWithRandomKeys(t *testing.T) {
	seedRand()
	for i := 0; i < *testRound; i++ {
		randomTest(t)
	}
}

func seedRand() {
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)
	fmt.Printf("seed=%d\n", *seed)
}

func randomTest(t *testing.T) {
	var actions []string
	tree := NewRTree()
	dict := make(map[string]string)
	randomStrings := GetTestStrings()

	for i := 0; i < *actionCount; i++ {
		key := randomStrings[rand.Intn(len(randomStrings))]

		doRandomAction(&actions, key, tree, dict)
		if !isEqual(tree, dict) {
			t.Errorf("Tree is not identical to Map, seed: %d", seed)
			printActions(actions)
			printRTree(tree)
			printMap(dict)
		}
	}
}

func GetTestStrings() []string {
	var str []rune
	var randomStrings []string

	GetRandomKeys(&str, *treeWidth, *treeHeight, &randomStrings)
	print("\n\nrandom strings")
	for _, str := range randomStrings {
		print(str)
	}

	return randomStrings
}

// GetRandomKeys returns random strings in backtrack way
func GetRandomKeys(str *[]rune, depth int, width int, randomStrings *[]string) {
	if depth == 0 {
		return
	}

	length := len(*str) // backup length for restoring
	for i := 0; i < width; i++ {
		// append random string
		AppendRandomString(str)
		*randomStrings = append(*randomStrings, string(*str))
		GetRandomKeys(str, depth-1, width, randomStrings)
		// restore length
		*str = (*str)[0:length]
	}
}

// AppendRandomString appends at most *maxLen runes on the end of str
func AppendRandomString(str *[]rune) {
	charNum := 26

	// generate random string
	length := rand.Intn(*maxLen)
	for i := 0; i < length; i++ {
		c := rune(int('a') + rand.Intn(charNum))
		*str = append(*str, c)
	}
}

func doRandomAction(actions *[]string, key string, tree *RTree, dict map[string]string) {
	r := rand.Intn(100)

	if r < *insertRatio {
		if tree != nil {
			tree.Insert(key, key)
		}
		if dict != nil {
			dict[key] = key
		}
		*actions = append(*actions, fmt.Sprintf("%s %s", insertAction, key))
	} else {
		if tree != nil {
			tree.Remove(key)
		}
		if dict != nil {
			delete(dict, key)
		}
		*actions = append(*actions, fmt.Sprintf("%s %s", removeAction, key))
	}
}

func isEqual(tree *RTree, dict map[string]string) bool {
	// check if all keys in map are also in radix tree
	for key := range dict {
		_, ok := tree.Get(key)
		if !ok {
			return false
		}
	}

	// check if all keys in rtree are also in map
	return preOrderAndCompare(tree.root, dict)
}

func preOrderAndCompare(n *node, M map[string]string) bool {
	if n == nil {
		return true
	}

	leaf, child, next := true, true, true
	if n.Leaf != nil {
		key := n.Leaf.Val.(string)
		_, ok := M[key]
		leaf = ok
	}
	if n.Children != nil {
		child = preOrderAndCompare(n.Children, M)
	}
	if n.Next != nil {
		next = preOrderAndCompare(n.Next, M)
	}

	if !leaf || !child || !next {
		print("false node")
		printNode(n)
		print(fmt.Sprintf("%t %t %t", leaf, child, next))
	}

	return leaf && child && next
}
