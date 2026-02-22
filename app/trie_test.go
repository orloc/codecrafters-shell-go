package main

import (
	"testing"
)

func TestTrieInsertAndFind(t *testing.T) {
	tests := []struct {
		name    string
		words   []string
		prefix  string
		want    []string
	}{
		{
			name:   "empty trie returns nil",
			words:  nil,
			prefix: "a",
			want:   nil,
		},
		{
			name:   "no match",
			words:  []string{"cat", "car", "card"},
			prefix: "z",
			want:   nil,
		},
		{
			name:   "single match",
			words:  []string{"echo", "exit", "type"},
			prefix: "ec",
			want:   []string{"echo"},
		},
		{
			name:   "multiple matches sorted",
			words:  []string{"cat", "car", "card"},
			prefix: "ca",
			want:   []string{"car", "card", "cat"},
		},
		{
			name:   "exact word is a match",
			words:  []string{"go", "gopher"},
			prefix: "go",
			want:   []string{"go", "gopher"},
		},
		{
			name:   "empty prefix returns all words sorted",
			words:  []string{"banana", "apple", "cherry"},
			prefix: "",
			want:   []string{"apple", "banana", "cherry"},
		},
		{
			name:   "overlapping prefixes",
			words:  []string{"test", "testing", "tested", "tester"},
			prefix: "test",
			want:   []string{"test", "tested", "tester", "testing"},
		},
		{
			name:   "single character prefix",
			words:  []string{"a", "ab", "abc", "b"},
			prefix: "a",
			want:   []string{"a", "ab", "abc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := newTrie()
			for _, w := range tt.words {
				tr.Insert(w)
			}
			got := tr.FindByPrefix(tt.prefix)
			if len(got) != len(tt.want) {
				t.Fatalf("FindByPrefix(%q) returned %d results, want %d\n  got:  %v\n  want: %v",
					tt.prefix, len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FindByPrefix(%q)[%d] = %q, want %q", tt.prefix, i, got[i], tt.want[i])
				}
			}
		})
	}
}
