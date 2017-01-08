package qradix

type node struct {
	Prefix   string
	Children *node
	Next     *node
	Leaf     *leafNode
}

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
	Root *node
	Size int
}

func commonPrefixID(s1 string, s2 string) int {
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
	//return &RTree{Root: &node{}}
	return &RTree{Root: &node{}}
}

// Get returns a value according to path
// if no match is found, it returns (nil, false) instead
func (T *RTree) Get(path string) (interface{}, bool) {
	if len(path) == 0 {
		return nil, false
	}

	n := T.Root
	for {
		if n == nil {
			return nil, false
		}

		id := commonPrefixID(n.Prefix, path)
		if id == -1 {
			n = n.Next
			continue
		} else if id == len(n.Prefix)-1 && id < len(path)-1 {
			path = path[id+1:]
			n = n.Children
			continue
		} else if id == len(n.Prefix)-1 && id == len(path)-1 && n.Leaf != nil {
			return n.Leaf.Val, true
		}
		return nil, false
	}
}

// split splits node into two nodes
// node1's prefix is [0, id)
// node2's prefix is [id, len-1]
func split(n *node, id int) (*node, bool) {
	if n == nil || id <= 0 || id > len(n.Prefix)-1 {
		return nil, false
	}

	newNode := &node{Prefix: n.Prefix[id:]}
	newNode.Children = n.Children
	newNode.Leaf = n.Leaf
	n.Children = newNode
	n.Leaf = nil
	n.Prefix = n.Prefix[:id]
	return newNode, true
}

// Insert adds value on the leaf of path
// if path already exists, it will update the value and returns former value.
func (T *RTree) Insert(path string, val interface{}) (interface{}, bool) {
	if len(path) == 0 || T.Root == nil {
		return nil, false
	}

	n := T.Root
	pathSuffix := path
	for {
		id := commonPrefixID(n.Prefix, pathSuffix)
		if id == -1 {
			if n.Next != nil {
				n = n.Next
				continue
			}
			n.Next = newNode(pathSuffix, nil, nil, &leafNode{Key: path, Val: val})
			T.Size++
			return nil, true
		} else if id < len(n.Prefix)-1 {
			childNode, ok := split(n, id+1)
			// TODO seems should be deleted
			if !ok {
				return nil, false
			}
			if id < len(pathSuffix)-1 {
				childNode.Next = newNode(pathSuffix[id+1:], nil, nil, &leafNode{Key: path, Val: val})
				T.Size++
				return nil, true
			}
			n.Leaf = &leafNode{Key: path, Val: val}
			T.Size++
			return nil, true
		}
		if id < len(pathSuffix)-1 {
			if n.Children != nil {
				pathSuffix = pathSuffix[id+1:]
				n = n.Children
				continue
			}
			n.Children = newNode(pathSuffix[id+1:], nil, nil, &leafNode{Key: path, Val: val})
			T.Size++
			return nil, true
		}
		return T.updateLeafVal(n, path, val)
	}
}

// updateLeafVal updates attributes of a leafNode
// if node has no leaf, new leafNode will be assigned to the node
// *node n must exist or it will create a new node
func (T *RTree) updateLeafVal(n *node, key string, newVal interface{}) (interface{}, bool) {
	if n.Leaf == nil {
		n.Leaf = &leafNode{Key: key, Val: newVal}
		T.Size++
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
		return false
	}
	parent := T.Root
	child := T.Root
	for {
		if child == nil {
			return false
		}

		id := commonPrefixID(child.Prefix, path)
		if id == -1 {
			child = child.Next
			continue
		} else if id == len(child.Prefix)-1 && id < len(path)-1 {
			path = path[id+1:]
			parent = child
			child = child.Children
			continue
		} else if id == len(child.Prefix)-1 && id == len(path)-1 && child.Leaf != nil {
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
	if T.Size > 0 {
		T.Size--
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
	if parent == T.Root {
		previousChild = T.Root
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
