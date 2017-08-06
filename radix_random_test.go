package qradix

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

const logOff bool = true   // is log off
const runTestCount int = 5 // how many times will random test run
const maxLen int = 5       // max length of random string
const treeWidth int = 2    // how many random suffix will be append after node's string
const treeDepth int = 10   // how many times will random suffix be append to root node's string
const insertRatio = 60     // control the the ratio between insert action and remove action
const actionCount = 500    // how many actions will be applied on RTree and map
const insertAction = "insert"
const removeAction = "remove"

var seed int64 // seed can be set to some value to re-produce test failure

// TestWithRandomKeys starts a random test by comparing RTree and map
func TestWithRandomKeys(t *testing.T) {
	seedRand()
	for i := 0; i < runTestCount; i++ {
		randomTest(t)
	}
}

func seedRand() {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed)
	print(fmt.Sprintf("seed=%d", seed))
}

func randomTest(t *testing.T) {
	var actions []string
	tree := NewRTree()
	dict := make(map[string]string)
	randomStrings := getTestStrings()

	for i := 0; i < actionCount; i++ {
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

func getTestStrings() []string {
	var str []rune
	var randomStrings []string

	getRandomKeys(&str, treeWidth, treeDepth, &randomStrings)
	print("\n\nrandom strings")
	for _, str := range randomStrings {
		print(str)
	}

	return randomStrings
}

// getRandomKeys returns random strings in backtrack way
func getRandomKeys(str *[]rune, depth int, width int, randomStrings *[]string) {
	if depth == 0 {
		return
	}

	length := len(*str) // backup length for restoring
	for i := 0; i < width; i++ {
		// append random string
		appendRandomString(str)
		*randomStrings = append(*randomStrings, string(*str))
		getRandomKeys(str, depth-1, width, randomStrings)
		// restore length
		*str = (*str)[0:length]
	}
}

// appendRandomString appends at most maxLen runes on the end of str
func appendRandomString(str *[]rune) {
	charNum := 26

	// generate random string
	length := rand.Intn(maxLen)
	for i := 0; i < length; i++ {
		c := rune(int('a') + rand.Intn(charNum))
		*str = append(*str, c)
	}
}

func doRandomAction(actions *[]string, key string, tree *RTree, dict map[string]string) {
	r := rand.Intn(100)

	if r < insertRatio {
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
	return preOrderAndCompare(tree.Root, dict)
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
		print(fmt.Sprintf("%s %s %s", leaf, child, next))
	}

	return leaf && child && next
}

// BFS is breadth first traverse on the radix tree
func BFS(T *RTree, function func(*node)) {
	Q := make([]*node, 0)
	Q = append(Q, T.Root)

	for len(Q) > 0 {
		n := Q[0]
		Q = Q[1:]

		if n == nil {
			continue
		}
		if n.Children != nil {
			Q = append(Q, n.Children)
		}
		if n.Next != nil {
			Q = append(Q, n.Next)
		}
		function(n)
	}
}

func print(msg string) {
	if !logOff {
		fmt.Println(msg)
	}
}

func printActions(actions []string) {
	print("\n\naction list:")
	for id, action := range actions {
		print(fmt.Sprintf("action%d: %s", id, action))
	}
}

func printRTree(T *RTree) {
	print("\n\nrtree:")
	BFS(T, printNode)
}

func printMap(M map[string]string) {
	print("\n\nmap:")
	for _, val := range M {
		print(val)
	}
}

// printNode prints all members of a node
func printNode(n *node) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("[prefix: %s] ", n.Prefix))
	if n.Children != nil {
		buf.WriteString(fmt.Sprintf("[child: %s] ", n.Children.Prefix))
	}
	if n.Next != nil {
		buf.WriteString(fmt.Sprintf("[next: %s] ", n.Next.Prefix))
	}
	if n.Leaf != nil {
		buf.WriteString(fmt.Sprintf("[value(key): %s]", n.Leaf.Key))
	}

	print(buf.String())
}
