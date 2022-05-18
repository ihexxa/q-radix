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

type Node interface {
	Segment() string
	FirstChild() (Node, bool)
	NextNode() (Node, bool)
	Value() (interface{}, bool)
	Extra() (interface{}, bool)
}

// BFS is breadth first traverse on the radix tree
func BFS(T *RTree, apply func(Node)) {
	Q := make([]Node, 0)
	if T.root == nil {
		return
	}

	Q = append(Q, T.root)
	for len(Q) > 0 {
		n := Q[0]
		Q = Q[1:]

		firstChild, ok := n.FirstChild()
		if ok {
			Q = append(Q, firstChild)
		}
		nextNode, ok := n.NextNode()
		if ok {
			Q = append(Q, nextNode)
		}
		apply(n)
	}
}

func print(msg string) {
	if *debug {
		fmt.Println(msg)
	}
}

func printActions(actions []string) {
	print("\naction list:")
	for id, action := range actions {
		print(fmt.Sprintf("action%d: %s", id, action))
	}
}

func printRTree(T *RTree) {
	print("\nrtree:")
	BFS(T, PrintNode)
}

func printMap(M map[string]string) {
	print("\nmap:")
	for _, val := range M {
		print(fmt.Sprintf("(%s)", val))
	}
}

// PrintNode prints node's fields
func PrintNode(n Node) {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("[prefix: %s] ", n.Segment()))
	firstChild, ok := n.FirstChild()
	if ok {
		buf.WriteString(fmt.Sprintf("[child: %s] ", firstChild.Segment()))
	}
	nextNode, ok := n.NextNode()
	if ok {
		buf.WriteString(fmt.Sprintf("[next: %s] ", nextNode.Segment()))
	}
	value, ok := n.Value()
	if ok {
		buf.WriteString(fmt.Sprintf("[value(key): %s]", value))
	}
	extra, ok := n.Extra()
	if ok {
		buf.WriteString("[idx:")
		idx := extra.(map[rune]*node)
		for rune1, node := range idx {
			buf.WriteString(fmt.Sprintf("(%s->%s)", string(rune1), node.Segment()))
		}
		buf.WriteString("]")
	}

	fmt.Println(buf.String())
}
