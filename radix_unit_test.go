package qradix

import "testing"

func TestInsert(t *testing.T) {
	rTree := NewRTree()
	cases := [][]string{
		{"a", "a"},
		{"b", "b"},
		{"ab", "ab"},
		{"中", "中"},
		{"中文", "中文"},
	}

	for _, testCase := range cases {
		rTree.Insert(testCase[0], testCase[0])
		if _, ok := rTree.Get(testCase[1]); !ok {
			t.Errorf("insert key [%s] is not found", testCase[1])
		}
	}
	if rTree.Size != len(cases) {
		t.Error("size of radix tree is not correct")
	}
}
func TestRemove(t *testing.T) {
	rTree := NewRTree()
	values := []string{
		"a",
		"b",
		"aa",
		"ba",
		"bb",
		"ca",
		"cb",
		"caa",
		"da",
		"db",
		"daa",
		"dba",
	}
	cases := [][]string{
		{"aa", "aa"},
		{"ba", "ba"},
		{"ca", "ca"},
		{"db", "db"},
	}

	for _, val := range values {
		rTree.Insert(val, val)
	}

	for _, testCase := range cases {
		if !rTree.Remove(testCase[1]) {
			t.Errorf("insert key [%s] is not removed", testCase[1])
		}
	}
	if rTree.Size != len(values)-len(cases) {
		t.Error("size of radix tree is not correct")
	}
}
