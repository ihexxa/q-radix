package qradix

import (
	"testing"
)

func TestOperations(t *testing.T) {
	t.Run("test Insert", testInsert)
	t.Run("test Remove", testRemove)
	t.Run("test GetAllMatches", testGetAllMatches)
	t.Run("test GetLongestMatch", testGetLongestMatch)
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
	if rTree.Size() != len(cases) {
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
	if rTree.Size() != len(values)-len(cases) {
		t.Error("the size of radix tree is not correct")
	}
}

func testGetAllMatches(t *testing.T) {
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
		&TestCase{
			desc:    "found a match shorter than key",
			inserts: []string{"a", "ab", "abd"},
			get:     "abc",
			expect:  []string{"a", "ab"},
		},
	}

	for _, tc := range testCases {
		rTree := NewRTree()

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

func testGetLongestMatch(t *testing.T) {
	type TestCase struct {
		desc    string
		inserts []string
		get     string
		expect  []string
	}

	testCases := []*TestCase{
		&TestCase{
			desc:    "found a match as long as key",
			inserts: []string{"a", "ab", "ac", "abc", "abcd"},
			get:     "abc",
			expect:  []string{"abc"},
		},
		&TestCase{
			desc:    "found a match shorter than key",
			inserts: []string{"a", "ab", "abd"},
			get:     "abc",
			expect:  []string{"ab"},
		},
		&TestCase{
			desc:    "no match found",
			inserts: []string{"a", "ab", "abd"},
			get:     "c",
			expect:  []string{},
		},
	}

	for _, tc := range testCases {
		rTree := NewRTree()

		for _, insert := range tc.inserts {
			rTree.Insert(insert, insert)
		}

		match, found := rTree.GetLongestMatch(tc.get)
		if len(tc.expect) == 0 {
			if found {
				t.Errorf("GetLongestMatch(%s): expect no match but found one %s", tc.desc, match)
			} else {
				continue
			}
		} else {
			if tc.expect[0] != match {
				t.Errorf("GetLongestMatch(%s): got %s expect %s", tc.desc, match, tc.expect[0])
			} else {
				continue
			}
		}
	}
}
