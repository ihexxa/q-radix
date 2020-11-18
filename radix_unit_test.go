package qradix

import "testing"

func TestOperations(t *testing.T) {
	t.Run("test Insert", testInsert)
	t.Run("test Remove", testRemove)
	t.Run("test GetAllMatches", testGetAllMatches)
}

func testInsert(t *testing.T) {
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
			t.Errorf("the insert key [%s] is not found", testCase[1])
		}
	}
	if rTree.Size != len(cases) {
		t.Error("the size of radix tree is not correct")
	}
}

func testRemove(t *testing.T) {
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
		t.Error("the size of radix tree is not correct")
	}
}

func testGetAllMatches(t *testing.T) {
	rTree := NewRTree()
	type TestCase struct {
		desc    string
		inserts []string
		get     string
		expect  []string
	}

	testCases := []*TestCase{
		&TestCase{
			desc:    "match along 1 of 2 braches",
			inserts: []string{"a", "ab", "ac", "abc", "abcd"},
			get:     "abc",
			expect:  []string{"a", "ab", "abc"},
		},
		&TestCase{
			desc:    "single word",
			inserts: []string{"a"},
			get:     "a",
			expect:  []string{"a"},
		},
	}

	for _, tc := range testCases {
		for _, insert := range tc.inserts {
			rTree.Insert(insert, insert)
		}
		matches := rTree.GetAllMatches(tc.get)
		for i, match := range matches {
			strMatch := match.(string)
			if strMatch != tc.expect[i] {
				t.Errorf("GetAllMatches: got %s expect %s", strMatch, tc.expect[i])
			}
		}
	}
}
