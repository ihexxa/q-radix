package qradix

import "fmt"

var errImpossible = func(prefix1, prefix2 string) string {
	return fmt.Sprintf("the first rune of %s and %s must be same", prefix1, prefix2)
}

// node is a node of radix tree and it is not a leaf
type node struct {
	Prefix   string
	Children *node
	Next     *node
	Leaf     *leafNode
	// Idx finds the sibling with the first rune of the current key
	Idx map[rune]*node
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
}

// return common prefix's offset of s1 and s2, in byte
// s1[:offset] == s2[:offset]
func commonPrefixOffset(s1, s2 string) int {
	i := 0
	runes1, runes2 := []rune(s1), []rune(s2)
	for ; i < len(runes1) && i < len(runes2); i++ {
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
	return &RTree{root: &node{Idx: map[rune]*node{}}}
}

// Size returns the size of the tree
func (T *RTree) Size() int {
	return T.size
}

// Get returns a value according to the key
// if the key does not exist, it returns (nil, false)
func (T *RTree) Get(key string) (interface{}, bool) {
	if len(key) == 0 {
		if T.root.Leaf != nil {
			return T.root.Leaf.Val, true
		}
		return nil, false
	}

	var ok bool
	var rune1 rune
	var matchedNode *node
	node1 := T.root
	for {
		if node1 == nil {
			return nil, false
		}

		// try to find the matched node in this level
		// with the first rune of key
		rune1 = []rune(key)[0]
		matchedNode, ok = node1.Idx[rune1]
		if !ok {
			return nil, false
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
			return matchedNode.Leaf.Val, true
		}
		return nil, false
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
func (T *RTree) Insert(key string, val interface{}) (interface{}, bool) {
	if T.root == nil {
		return nil, false
	}
	if len(key) == 0 {
		return T.updateLeafVal(T.root, "", val)
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
			return nil, true
		}

		offset := commonPrefixOffset(matchedNode.Prefix, pathSuffix)
		if offset == -1 {
			// this is impossible
			panic(errImpossible(matchedNode.Prefix, key))

		} else if offset < len(matchedNode.Prefix)-1 {
			// partial matched to matchedNode.Prefix
			childNode, ok := split(matchedNode, offset+1)
			if !ok {
				return nil, false
			}
			// pathSuffix is longer, add the node as child's sibling
			if offset < len(pathSuffix)-1 {
				newNodePrefix := pathSuffix[offset+1:]
				childNode.Next = newNode(newNodePrefix, nil, nil, &leafNode{Val: val})
				childNode.Idx[[]rune(newNodePrefix)[0]] = childNode.Next
				T.size++
				return nil, true
			}
			// pathSuffix is same as n'prefix, update n's leaf
			// matchedNode must have no leaf because it was just splitted
			matchedNode.Leaf = &leafNode{Val: val}
			T.size++
			return nil, true
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
			return nil, true
		}

		// update current node's leaf
		return T.updateLeafVal(matchedNode, key, val)
	}
}

// updateLeafVal updates fields of a leafNode
// if node has no leaf, a new leafNode will be assigned to the node
// *node n must exist or it will create a new node
func (T *RTree) updateLeafVal(n *node, key string, newVal interface{}) (interface{}, bool) {
	if n.Leaf == nil {
		n.Leaf = &leafNode{Val: newVal}
		T.size++
		return nil, true
	}

	oldVal := n.Leaf.Val
	n.Leaf.Val = newVal
	return oldVal, true
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
	if len(key) == 0 {
		if T.root.Leaf != nil {
			T.root.Leaf = nil
			T.size--
			return true
		}
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
			return T.removeChild(parent, matchedNode)
		}
		return false
	}
}

// removeChild deletes child node from Tree T
func (T *RTree) removeChild(parent *node, child *node) bool {
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
		delete(child.Idx, []rune(child.Prefix)[0])
		if child.Next != nil {
			child.Next.Idx = child.Idx
		}
		parent.Children = child.Next
		return true
	}
	// child is not the first child
	// search for the previous node of child
	previousChild := parent.Children
	if parent == T.root {
		// when parent is T.root
		// it means parent and parent.Children are in the same level
		previousChild = T.root
	}
	delete(previousChild.Idx, []rune(child.Prefix)[0])

	for previousChild != nil && previousChild.Next != child {
		previousChild = previousChild.Next
	}
	if previousChild == nil {
		panic("the previousChild not found")
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
	resultMap := map[string]interface{}{}
	if T.root.Leaf != nil {
		resultMap[""] = T.root.Leaf.Val
	}
	if len(key) == 0 {
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

// GetBestMatch returns the longest match in the tree according to the key
// if there is no match, it returns empty string, nil and false
func (T *RTree) GetBestMatch(key string) (string, interface{}, bool) {
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
