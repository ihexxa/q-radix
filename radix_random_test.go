package qradix

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
)

const (
	insertAction = "insert"
	removeAction = "remove"
)

var (
	// use -args to input these options: "go test -args -h=10"
	actionCount = flag.Int("c", 30, "how many actions will be applied on RTree and map")
	insertRatio = flag.Int("i", 60, "control the the ratio between insert action and remove action")
	maxLen      = flag.Int("l", 5, "how long will random string be generated")
	seed        = flag.Int64("s", 0, "seed can be set to specific value to re-produce test failure")
	testRound   = flag.Int("r", 10, "how many times will random test run")
	treeHeight  = flag.Int("h", 6, "how many times will random suffix be append to root node's string")
	treeWidth   = flag.Int("w", 3, "how many random suffix will be append after node's string")
)

// TestWithRandomKeys starts a random test by comparing a RTree and a map
func TestWithRandomKeys(t *testing.T) {
	seedRand()
	for i := 0; i < *testRound; i++ {
		randomTest(t)
		longerKeyTest(t)
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
			printActions(actions)
			printRTree(tree)
			printMap(dict)
			t.Fatalf("Tree is not identical to Map, seed: %d", *seed)
		}
		if tree.Size() != len(dict) {
			printActions(actions)
			printRTree(tree)
			printMap(dict)
			t.Fatalf(
				"incorrect size: got(%d) expected(%d) (seed: %d)",
				tree.Size(),
				len(dict),
				*seed,
			)
		}
		prefixes := tree.GetAllPrefixMatches(key)
		if !checkPrefixMatches(key, prefixes, tree, dict) {
			fmt.Printf("prefixes of (%s): %+v\n", key, prefixes)
			printActions(actions)
			printRTree(tree)
			printMap(dict)
			t.Fatalf("incorrect prefix matches")
		}

		// test marshaling
		tree2 := NewRTree()
		rows := []string{}
		for row := range tree.String() {
			rows = append(rows, row)
		}
		// size of chan is known
		inputChan := make(chan string, len(rows))
		go func() {
			for _, row := range rows {
				inputChan <- row
			}
			close(inputChan)
		}()
		err := tree2.FromString(inputChan)
		if err != nil {
			printActions(actions)
			printRTree(tree2)
			fmt.Println("rows")
			for _, row := range rows {
				fmt.Println(row)
			}
			printMap(dict)
			t.Fatal(err)
		}
		if !isEqual(tree2, dict) {
			printActions(actions)
			printRTree(tree2)
			printMap(dict)
			t.Fatalf("Tree2 is not identical to Map, seed: %d", *seed)
		}
		if tree2.Size() != len(dict) {
			t.Fatalf(
				"incorrect tree2 size: got(%d) expected(%d) (seed: %d)",
				tree2.Size(),
				len(dict),
				*seed,
			)
		}
	}
}

func longerKeyTest(t *testing.T) {
	var actions []string
	tree := NewRTree()
	dict := make(map[string]string)
	randomStrings := GetTestStrings()

	for i := 0; i < *actionCount; i++ {
		key := randomStrings[rand.Intn(len(randomStrings))]
		doRandomAction(&actions, key, tree, dict)
	}

	randomKey := randomStrings[rand.Intn(len(randomStrings))]
	end := rand.Intn(len(randomKey)) + 1
	start := rand.Intn(end)
	randomKey = randomKey[start:end]
	longerMatches := tree.GetLongerMatches(randomKey, 5)
	if !checkLongerMatches(randomKey, longerMatches) {
		fmt.Printf("longerMatches of (%s): %+v\n", randomKey, longerMatches)
		printActions(actions)
		printRTree(tree)
		printMap(dict)
		t.Fatalf("incorrect longer matches")
	}
}

func checkPrefixMatches(
	key string,
	prefixes map[string]interface{},
	tree *RTree,
	dict map[string]string,
) bool {
	for offset := range []rune(key) {
		prefix := string([]rune(key)[:offset])
		val1, ok := dict[prefix]
		if ok {
			val2, ok2 := prefixes[prefix]
			if !ok2 || val1 != val2.(string) {
				fmt.Printf("prefix map(%s) tree(%+v) not match\n", val1, val2)
				return false
			}
		}
	}
	return true
}

// checkLongerMatches only checks if key is the prefix of each longerMatches
// Becasue currently, besides the first one,
// it doesn't return results by the order of how much it is closed to the key
func checkLongerMatches(
	key string,
	longerMatches map[string]interface{},
) bool {
	for longerMatch := range longerMatches {
		if !strings.HasPrefix(longerMatch, key) {
			fmt.Printf("%s is not a prefix of %s \n", key, longerMatch)
			return false
		}
	}

	return true
}

func GetTestStrings() []string {
	var str []rune
	var randomStrings []string

	GetRandomKeys(&str, *treeWidth, *treeHeight, &randomStrings)
	// print("\n\nrandom strings")
	// for _, str := range randomStrings {
	// 	print(str)
	// }

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
	length := rand.Intn(*maxLen) + 1 // empty string is not allowed
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
		*actions = append(*actions, fmt.Sprintf("%s (%s)", insertAction, key))
	} else {
		if tree != nil {
			tree.Remove(key)
		}
		if dict != nil {
			delete(dict, key)
		}
		*actions = append(*actions, fmt.Sprintf("%s (%s)", removeAction, key))
	}
}

func isEqual(tree *RTree, dict map[string]string) bool {
	// check if all keys in map are also in radix tree
	for key := range dict {
		_, err := tree.Get(key)
		if err != nil {
			fmt.Printf("get key(%s) should be ok (%s)\n", key, err)
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
		PrintNode(n)
		print(fmt.Sprintf("%t %t %t", leaf, child, next))
	}

	return leaf && child && next
}
