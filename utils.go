package qradix

import (
	"bytes"
	"flag"
	"fmt"
)

var (
	// enable debug mode by "go test -args -d=true"
	debug = flag.Bool("d", false, "print debug messages")
)

// BFS is breadth first traverse on the radix tree
func BFS(T *RTree, function func(*node)) {
	Q := make([]*node, 0)
	Q = append(Q, T.root)

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
	if *debug {
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
		buf.WriteString(fmt.Sprintf("[value(key): %s]", n.Leaf.Val))
	}
	if n.Idx != nil {
		buf.WriteString("[idx:")
		for rune1, node := range n.Idx {
			buf.WriteString(fmt.Sprintf("(%s->%s)", string(rune1), node.Prefix))
		}
		buf.WriteString("]")
	}

	print(buf.String())
}
