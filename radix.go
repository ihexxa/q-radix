package qradix

// node is a node of radix tree and it is not a leaf
type node struct {
	Prefix   string
	Children *node
	Next     *node
	Leaf     *leafNode
}

// leafNode stores all values
type leafNode struct {
	Key string
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

func commonPrefixOffset(s1 string, s2 string) int {
	i := 0
	for ; i < len(s1) && i < len(s2); i++ {
		if s1[i] != s2[i] {
			break
		}
	}
	return i - 1
}

// NewRTree returns a new radix tree
func NewRTree() *RTree {
	return &RTree{root: &node{}}
}

// Size returns the size of the tree
func (T *RTree) Size() int {
	return T.size
}

// Get returns a value according to key
// if no match is found, it returns (nil, false) instead
func (T *RTree) Get(key string) (interface{}, bool) {
	if len(key) == 0 {
		if T.root.Leaf != nil {
			return T.root.Leaf.Val, true
		}
		return nil, false
	}

	n := T.root
	for {
		if n == nil {
			return nil, false
		}

		offset := commonPrefixOffset(n.Prefix, key)
		if offset == -1 {
			n = n.Next
			continue
		} else if offset == len(n.Prefix)-1 && offset < len(key)-1 {
			key = key[offset+1:]
			n = n.Children
			continue
		} else if offset == len(n.Prefix)-1 && offset == len(key)-1 && n.Leaf != nil {
			return n.Leaf.Val, true
		}
		return nil, false
	}
}

// GetAllMatches returns all the prefix matches in the tree according to the key
// if no match is found, it returns an empty slice
func (T *RTree) GetAllMatches(key string) []interface{} {
	results := []interface{}{}
	if len(key) == 0 {
		if T.root.Leaf != nil {
			results = append(results, T.root.Leaf.Val)
			return results
		}
		return results
	}

	n := T.root
	for {
		if n == nil {
			return results
		}

		offset := commonPrefixOffset(n.Prefix, key)
		if offset == -1 {
			n = n.Next
			continue
		} else if offset == len(n.Prefix)-1 && offset < len(key)-1 {
			if n.Leaf != nil {
				results = append(results, n.Leaf.Val)
			}
			key = key[offset+1:]
			n = n.Children
			continue
		} else if offset == len(n.Prefix)-1 && offset == len(key)-1 && n.Leaf != nil {
			results = append(results, n.Leaf.Val)
			return results
		}
		return results
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
	n.Children = newNode
	n.Leaf = nil
	n.Prefix = n.Prefix[:offset]
	return newNode, true
}

// Insert adds a value in the tree. The value can be found by the key.
// if path already exists, it updates the value and returns former value.
func (T *RTree) Insert(key string, val interface{}) (interface{}, bool) {
	if T.root == nil {
		return nil, false
	}
	if len(key) == 0 {
		return T.updateLeafVal(T.root, "", val)
	}

	n := T.root
	pathSuffix := key
	for {
		offset := commonPrefixOffset(n.Prefix, pathSuffix)
		if offset == -1 {
			if n.Next != nil {
				n = n.Next
				continue
			}
			n.Next = newNode(pathSuffix, nil, nil, &leafNode{Key: key, Val: val})
			T.size++
			return nil, true
		} else if offset < len(n.Prefix)-1 {
			childNode, ok := split(n, offset+1)
			if !ok {
				return nil, false
			}
			if offset < len(pathSuffix)-1 {
				childNode.Next = newNode(pathSuffix[offset+1:], nil, nil, &leafNode{Key: key, Val: val})
				T.size++
				return nil, true
			}
			n.Leaf = &leafNode{Key: key, Val: val}
			T.size++
			return nil, true
		}
		if offset < len(pathSuffix)-1 {
			if n.Children != nil {
				pathSuffix = pathSuffix[offset+1:]
				n = n.Children
				continue
			}
			n.Children = newNode(pathSuffix[offset+1:], nil, nil, &leafNode{Key: key, Val: val})
			T.size++
			return nil, true
		}
		return T.updateLeafVal(n, key, val)
	}
}

// updateLeafVal updates fields of a leafNode
// if node has no leaf, a new leafNode will be assigned to the node
// *node n must exist or it will create a new node
func (T *RTree) updateLeafVal(n *node, key string, newVal interface{}) (interface{}, bool) {
	if n.Leaf == nil {
		n.Leaf = &leafNode{Key: key, Val: newVal}
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

// Remove deletes leaf node at the end of the path
// if the leaf node doesn't exist, it will return false
func (T *RTree) Remove(path string) bool {
	if len(path) == 0 {
		if T.root.Leaf != nil {
			T.root.Leaf = nil
			return true
		}
		return false
	}

	parent := T.root
	child := T.root
	for {
		if child == nil {
			return false
		}

		offset := commonPrefixOffset(child.Prefix, path)
		if offset == -1 {
			child = child.Next
			continue
		} else if offset == len(child.Prefix)-1 && offset < len(path)-1 {
			path = path[offset+1:]
			parent = child
			child = child.Children
			continue
		} else if offset == len(child.Prefix)-1 && offset == len(path)-1 && child.Leaf != nil {
			return T.removeChild(parent, child)
		}
		return false
	}
}

// removeChild deletes child node from Tree T
func (T *RTree) removeChild(parent *node, child *node) bool {
	if child == nil {
		return false
	}

	child.Leaf = nil
	if T.size > 0 {
		T.size--
	}
	if child.Children != nil {
		if child.Next == nil {
			merge(child, child.Children)
		}
		return true
	}
	// child is the first child and has no child
	if parent.Children == child {
		parent.Children = child.Next
		return true
	}
	// child is not the first child, search for the previous node of child
	previousChild := parent.Children
	if parent == T.root {
		previousChild = T.root
	}
	for previousChild != nil && previousChild.Next != child {
		previousChild = previousChild.Next
	}
	if previousChild == nil {
		return false
	}
	previousChild.Next = previousChild.Next.Next
	// if only 1 child is left
	merge(parent, parent.Children)
	return true
}
