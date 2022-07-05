package qradix

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrEmptyKey     = errors.New("empty key is not allowed")
	ErrNotExist     = errors.New("key not exist")
	ErrInvalidSplit = errors.New("invalid split")
	errImpossible   = func(prefix1, prefix2 string) string {
		return fmt.Sprintf("the first rune of %s and %s must be same", prefix1, prefix2)
	}
)

// node is a node of radix tree and it is not a leaf
type node struct {
	Prefix   string
	Children *node
	Next     *node
	Leaf     *leafNode
	// Idx finds the sibling with the first rune of the current key
	Idx map[rune]*node
}

// Segment returns node's segment
func (n *node) Segment() string {
	return n.Prefix
}

// FirstChild returns node's first child, it returns nil if there is no
func (n *node) FirstChild() (Node, bool) {
	return n.Children, n.Children != nil
}

// NextNode returns node's next node, it returns nil if there is no
func (n *node) NextNode() (Node, bool) {
	return n.Next, n.Next != nil
}

// Value returns node's value, it returns nil if there is no
func (n *node) Value() (interface{}, bool) {
	if n.Leaf != nil {
		return n.Leaf.Val, true
	}
	return nil, false
}

// Extra returns node's Extra Info
func (n *node) Extra() (interface{}, bool) {
	return n.Idx, n.Idx != nil
}

// leafNode stores all values
type leafNode struct {
	Val interface{}
}

func newNode(prefix string, children *node, next *node, leaf *leafNode) *node {
	return &node{
		Prefix:   prefix,
		Children: children,
		Next:     next,
		Leaf:     leaf,
	}
}

// RTree is a radix tree
type RTree struct {
	root *node
	size int
	m    *sync.RWMutex
}

// return common prefix's offset of s1 and s2, in byte
// s1[:offset+1] == s2[:offset+1]
func commonPrefixOffset(s1, s2 string) int {
	i := 0
	runes1, runes2 := []rune(s1), []rune(s2)
	length := len(runes1)
	if len(runes2) < length {
		length = len(runes2)
	}
	for ; i < length; i++ {
		if runes1[i] != runes2[i] {
			break
		}
	}

	if i == 0 {
		return -1
	}
	return len(string(runes1[:i])) - 1
}

// NewRTree returns a new radix tree
func NewRTree() *RTree {
	return &RTree{
		root: nil,
		m:    &sync.RWMutex{},
	}
}

// Size returns the size of the tree
func (T *RTree) Size() int {
	T.m.RLock()
	defer T.m.RUnlock()
	return T.size
}

// Get returns a value according to the key
// if the key does not exist, it returns (nil, false)
func (T *RTree) Get(key string) (interface{}, error) {
	T.m.RLock()
	defer T.m.RUnlock()

	if len(key) == 0 {
		return nil, ErrEmptyKey
	}

	var ok bool
	var rune1 rune
	var matchedNode *node
	node1 := T.root
	for {
		if node1 == nil {
			return nil, ErrNotExist
		}

		// try to find the matched node in this level
		// with the first rune of key
		rune1 = []rune(key)[0]
		matchedNode, ok = node1.Idx[rune1]
		if !ok {
			return nil, ErrNotExist
		}

		offset := commonPrefixOffset(matchedNode.Prefix, key)
		if offset == -1 {
			// this is impossible
			panic(errImpossible(matchedNode.Prefix, key))
		} else if offset == len(matchedNode.Prefix)-1 && offset < len(key)-1 {
			key = key[offset+1:]
			node1 = matchedNode.Children
			continue
		} else if offset == len(matchedNode.Prefix)-1 &&
			offset == len(key)-1 &&
			matchedNode.Leaf != nil {
			return matchedNode.Leaf.Val, nil
		}
		return nil, ErrNotExist
	}
}

// split splits node into two nodes: parent and child.
// node1's prefix is [0, offset)
// node2's prefix is [offset, len-1]
func split(n *node, offset int) (*node, bool) {
	if n == nil || offset <= 0 || offset > len(n.Prefix)-1 {
		return nil, false
	}

	newNode := &node{Prefix: n.Prefix[offset:]}
	newNode.Children = n.Children
	newNode.Leaf = n.Leaf
	newNode.Idx = map[rune]*node{
		[]rune(newNode.Prefix)[0]: newNode, // add self to index
	}
	n.Children = newNode
	n.Leaf = nil
	n.Prefix = n.Prefix[:offset]
	return newNode, true
}

// Insert adds a value in the tree. Then the value can be found by the key.
// if path already exists, it updates the value and returns the former value.
func (T *RTree) Insert(key string, val interface{}) (interface{}, error) {
	T.m.Lock()
	defer T.m.Unlock()

	if len(key) == 0 {
		return nil, ErrEmptyKey
	}
	if T.root == nil {
		T.root = &node{
			Prefix: key,
			Leaf:   &leafNode{Val: val},
			Idx:    map[rune]*node{},
		}
		T.root.Idx[getRune1(key)] = T.root
		T.size = 1
		return nil, nil
	}

	pathSuffix := key
	// node1 is the first node at this level
	// matchedNode is the node
	// which first rune matches to the first rune of key
	var ok bool
	var rune1 rune
	var matchedNode *node
	var node1 = T.root
	for {
		// search the key level by level
		rune1 = []rune(pathSuffix)[0]
		matchedNode, ok = node1.Idx[rune1]
		if !ok {
			// no match in this level, insert a new node after the node1
			newNode := newNode(pathSuffix, nil, nil, &leafNode{Val: val})
			newNode.Next = node1.Next
			node1.Next = newNode
			node1.Idx[rune1] = newNode
			T.size++
			return nil, nil
		}

		offset := commonPrefixOffset(matchedNode.Prefix, pathSuffix)
		if offset == -1 {
			// this is impossible
			panic(errImpossible(matchedNode.Prefix, key))

		} else if offset < len(matchedNode.Prefix)-1 {
			// partial matched to matchedNode.Prefix
			childNode, ok := split(matchedNode, offset+1)
			if !ok {
				return nil, ErrInvalidSplit
			}
			// pathSuffix is longer, add the node as child's sibling
			if offset < len(pathSuffix)-1 {
				newNodePrefix := pathSuffix[offset+1:]
				childNode.Next = newNode(newNodePrefix, nil, nil, &leafNode{Val: val})
				childNode.Idx[[]rune(newNodePrefix)[0]] = childNode.Next
				T.size++
				return nil, nil
			}
			// pathSuffix is same as n'prefix, update n's leaf
			// matchedNode must have no leaf because it was just splitted
			matchedNode.Leaf = &leafNode{Val: val}
			T.size++
			return nil, nil
		}
		if offset < len(pathSuffix)-1 {
			// search children for left pathSuffix
			if matchedNode.Children != nil {
				pathSuffix = pathSuffix[offset+1:]
				node1 = matchedNode.Children
				continue
			}
			// matchedNode has no children, add the first child with pathSuffix[offset+1:]
			newNodePrefix := pathSuffix[offset+1:]
			matchedNode.Children = newNode(newNodePrefix, nil, nil, &leafNode{Val: val})
			matchedNode.Children.Idx = map[rune]*node{}
			matchedNode.Children.Idx[[]rune(newNodePrefix)[0]] = matchedNode.Children
			T.size++
			return nil, nil
		}

		// update current node's leaf
		return T.updateLeafVal(matchedNode, key, val)
	}
}

// updateLeafVal updates fields of a leafNode
// if node has no leaf, a new leafNode will be assigned to the node
// *node n must exist or it will create a new node
func (T *RTree) updateLeafVal(n *node, key string, newVal interface{}) (interface{}, error) {
	if n.Leaf == nil {
		n.Leaf = &leafNode{Val: newVal}
		T.size++
		return nil, nil
	}

	oldVal := n.Leaf.Val
	n.Leaf.Val = newVal
	return oldVal, nil
}

// merge merges parent node and parent's first child node
func merge(parent *node, child *node) bool {
	if parent != nil &&
		parent.Children != nil &&
		parent.Children == child &&
		parent.Leaf == nil &&
		parent.Next == nil &&
		child.Next == nil {

		parent.Prefix = parent.Prefix + child.Prefix
		parent.Leaf = child.Leaf
		parent.Children = child.Children
		return true
	}
	return false
}

// Remove deletes the leaf node according to the path
// if the leaf node exists, it will be deleted and "true" will be returned.
// or "false" will be returned.
func (T *RTree) Remove(key string) bool {
	T.m.Lock()
	defer T.m.Unlock()

	if len(key) == 0 {
		return false
	}

	// TODO: it is a little confuse here
	// because at the root level, parent is actually a sibling of the child, not parent
	pathSuffix := key
	parent := T.root
	node1 := T.root
	var matchedNode *node
	var ok bool
	var rune1 rune
	for {
		if node1 == nil {
			return false
		}
		rune1 = []rune(pathSuffix)[0]
		matchedNode, ok = node1.Idx[rune1]
		if !ok {
			// no match at this level
			return false
		}

		offset := commonPrefixOffset(matchedNode.Prefix, pathSuffix)
		if offset == -1 {
			// this is impossible
			panic(errImpossible(matchedNode.Prefix, pathSuffix))
		} else if offset == len(matchedNode.Prefix)-1 && offset < len(pathSuffix)-1 {
			pathSuffix = pathSuffix[offset+1:]
			parent = matchedNode
			node1 = matchedNode.Children
			continue
		} else if offset == len(matchedNode.Prefix)-1 &&
			offset == len(pathSuffix)-1 &&
			matchedNode.Leaf != nil {
			return T.removeChild(parent, matchedNode, parent == node1)
		}
		return false
	}
}

// removeChild deletes child node from Tree T
func (T *RTree) removeChild(parent *node, child *node, isParentSameLevel bool) bool {
	if child == nil {
		return false
	}
	if child.Leaf != nil {
		child.Leaf = nil
		T.size--
	}

	// if child has no sibling
	// merge it with its children
	if child.Children != nil {
		merge(child, child.Children)
		return true
	}
	// child is the first child
	// and it has no child， delete child
	if parent.Children == child {
		delete(child.Idx, getRune1(child.Prefix))
		if child.Next != nil {
			child.Next.Idx = child.Idx
		}
		parent.Children = child.Next
		return true
	}
	// child is not the first child
	// search for the previous node of child
	previousChild := parent.Children
	if isParentSameLevel {
		if parent == child {
			// delete the first node at the first level
			if parent.Next != nil {
				delete(parent.Idx, getRune1(parent.Prefix))
				parent.Next.Idx = parent.Idx
			}
			T.root = parent.Next
			return true
		} else {
			previousChild = parent
		}
	}
	delete(previousChild.Idx, []rune(child.Prefix)[0])

	for previousChild != nil && previousChild.Next != child {
		previousChild = previousChild.Next
	}
	if previousChild == nil {
		panic(fmt.Sprintf("the previousChild not found parent(%+v) previous(%+v) child(%+v)", parent, previousChild, child))
	}
	previousChild.Next = previousChild.Next.Next
	// merge will try to merge parent
	// and parent's first child if there is only 1 child left
	merge(parent, parent.Children)
	return true
}

func getRune1(key string) rune {
	return []rune(key)[0]
}

// GetAllPrefixMatches returns all prefix matches in the tree according to the key
// if no match is found, it returns an empty map
func (T *RTree) GetAllPrefixMatches(key string) map[string]interface{} {
	T.m.RLock()
	defer T.m.RUnlock()

	resultMap := map[string]interface{}{}
	if T.root == nil {
		return resultMap
	} else if len(key) == 0 {
		return resultMap
	}

	var ok bool
	var rune1 rune
	var matchedNode *node
	node1 := T.root
	pathSuffix := key
	baseOffset := 0 // key[:baseOffset+1] is matched
	for {
		if node1 == nil {
			break
		}

		rune1 = getRune1(pathSuffix)
		matchedNode, ok = node1.Idx[rune1]
		if !ok {
			break
		}

		offset := commonPrefixOffset(matchedNode.Prefix, pathSuffix)
		if offset == -1 {
			// this is impossible
			panic(errImpossible(matchedNode.Prefix, key))
		} else if offset < len(matchedNode.Prefix)-1 {
			break
		}
		if matchedNode.Leaf != nil {
			resultMap[key[:baseOffset+offset+1]] = matchedNode.Leaf.Val
		}

		if offset == len(pathSuffix)-1 {
			break
		}
		pathSuffix = pathSuffix[offset+1:]
		node1 = matchedNode.Children
		baseOffset += offset + 1
		continue
	}

	return resultMap
}

type traverseLog struct {
	n    *node
	base string
}

// GetLongerMatches returns at most `limmit` matches which are longer than the key
// if no match is found, it returns an empty map
func (T *RTree) GetLongerMatches(key string, limit int) map[string]interface{} {
	T.m.RLock()
	defer T.m.RUnlock()

	resultMap := map[string]interface{}{}
	if T.root == nil {
		return resultMap
	} else if len(key) == 0 {
		return resultMap
	}

	var ok bool
	var rune1 rune
	var matchedNode *node
	node1 := T.root
	pathSuffix := key
	baseOffset := 0 // key[:baseOffset] is matched
	for {
		if node1 == nil {
			return resultMap
		}

		rune1 = getRune1(pathSuffix)
		matchedNode, ok = node1.Idx[rune1]
		if !ok {
			return resultMap
		}

		offset := commonPrefixOffset(matchedNode.Prefix, pathSuffix)
		if offset == -1 {
			// this is impossible
			panic(errImpossible(matchedNode.Prefix, key))
		} else if offset == len(matchedNode.Prefix)-1 && offset < len(pathSuffix)-1 {
			pathSuffix = pathSuffix[offset+1:]
			node1 = matchedNode.Children
			baseOffset += offset + 1
			continue
		} else if offset == len(pathSuffix)-1 {
			break
		}
		return resultMap
	}

	if matchedNode.Leaf != nil {
		resultMap[key[:baseOffset]+matchedNode.Prefix] = matchedNode.Leaf.Val
	}
	// start from next level becasue matchedNode's siblings are not results
	// traverse from the matchedNode and return values
	queue := []*traverseLog{&traverseLog{
		n:    matchedNode.Children,
		base: key[:baseOffset] + matchedNode.Prefix,
	}}
	for len(queue) > 0 {
		tlog := queue[0]
		queue = queue[1:]
		if tlog.n == nil {
			continue
		}

		if tlog.n.Leaf != nil {
			resultMap[tlog.base+tlog.n.Prefix] = tlog.n.Leaf.Val
			if len(resultMap) > limit {
				break
			}
		}
		if tlog.n.Next != nil {
			queue = append(queue, &traverseLog{
				n:    tlog.n.Next,
				base: tlog.base,
			})
		}
		if tlog.n.Children != nil {
			queue = append(queue, &traverseLog{
				n:    tlog.n.Children,
				base: tlog.base + tlog.n.Prefix,
			})
		}
	}

	return resultMap
}

// GetBestMatch returns the longest match from all existings values which key is short than the input key
// if there is no match, it returns empty string, nil and false
func (T *RTree) GetBestMatch(key string) (string, interface{}, bool) {
	T.m.RLock()
	defer T.m.RUnlock()

	matches := T.GetAllPrefixMatches(key)
	if len(matches) == 0 {
		return "", nil, false
	}
	bestPrefix := ""
	for prefix := range matches {
		if len(prefix) > len(bestPrefix) {
			bestPrefix = prefix
		}
	}
	return bestPrefix, matches[bestPrefix], true
}

type visitLog struct {
	visited bool // if the node's value is already logged
	node    *node
	indents int // current indents
}

func intoRow(indents int, prefix, value string) string {
	return fmt.Sprintf(
		"%s%s\t\t%s",
		strings.Repeat("\t", indents),
		strings.ReplaceAll(prefix, "\t", "+\t"),
		strings.ReplaceAll(value, "\t", "+\t"),
	)
}

func fromRow(row string) (int, string, string) {
	indents := 0
	runes := []rune(row)
	for i, r := range runes {
		if string(r) != "\t" {
			indents = i
			break
		}
	}

	sepPos := 0
	keyAndValue := runes[indents:]
	for i := 0; i < len(keyAndValue)-1; i++ {
		if string(keyAndValue[i]) == "\t" && string(keyAndValue[i+1]) == "\t" {
			sepPos = i
			break
		}
	}

	return indents,
		strings.ReplaceAll(string(keyAndValue[:sepPos]), "+\t", "\t"),
		strings.ReplaceAll(string(keyAndValue[sepPos+2:]), "+\t", "\t")
}

// String serializes nodes one by one and sends them to channel in order.
// NOTICE: only string value is supported, or it will panic.
func (T *RTree) String() chan string {
	results := make(chan string, 512)
	if T.root == nil {
		close(results)
		return results
	}

	stack := make([]*visitLog, 0)
	stack = append(stack, &visitLog{
		node:    T.root,
		visited: false,
		indents: 0,
	})

	worker := func() {
		defer close(results)

		for len(stack) > 0 {
			vlog := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if !vlog.visited {
				// prefix is always logged (for restoring) even there is no leaf
				value := ""
				if vlog.node.Leaf != nil {
					value = vlog.node.Leaf.Val.(string) // or it will panic
				}
				results <- intoRow(
					vlog.indents,
					vlog.node.Prefix,
					value,
				)

				if vlog.node.Children != nil {
					// push vlog back
					vlog.visited = true
					stack = append(stack, vlog)

					// push the first child
					stack = append(stack, &visitLog{
						node:    vlog.node.Children,
						visited: false,
						indents: vlog.indents + 1,
					})

					continue
				}
			}
			if vlog.node.Next != nil {
				stack = append(stack, &visitLog{
					node:    vlog.node.Next,
					visited: false,
					indents: vlog.indents,
				})
			}
		}
	}
	go worker()

	return results
}

// FromString gets rows(nodes) from channel in order and add them to tree one by one.
// NOTICE:
// 1. only string value is supported, or it will panic.
// 2. The order of rows must be exactly same as String()'s output.
func (T *RTree) FromString(input chan string) error {
	parentsStack := []string{}
	for row := range input {
		indents, prefix, val := fromRow(row)

		if len(parentsStack) > 0 {
			if len(parentsStack) == indents {
				// previous row is the parent of this node,
				parentsStack = append(parentsStack, prefix)
			} else if len(parentsStack) > indents {
				// previous row is child of a row above
				parentsStack = parentsStack[:indents]
				parentsStack = append(parentsStack, prefix)
			} else {
				return fmt.Errorf("invalid indent previous(%d) current(%d)", len(parentsStack), indents)
			}
		} else {
			if indents > 0 {
				return fmt.Errorf("invalid indent previous(0) current(%d)", indents)
			} else {
				parentsStack = append(parentsStack, prefix)
			}
		}

		fullPrefix := strings.Join(parentsStack, "")
		if val != "" {
			_, err := T.Insert(fullPrefix, val)
			if err != nil {
				return fmt.Errorf("inserting error: %w", err)
			}
		}
	}

	return nil
}
