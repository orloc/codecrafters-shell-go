// trie.go — prefix trie for fast command name lookup.
//
// Used by the TAB completer to find all commands sharing a common prefix.
// Insert is O(k) where k is word length; FindByPrefix collects all
// descendants and returns them sorted.
package main

import "sort"

// trieNode is a single node in the prefix trie. Each node maps runes to
// children and marks whether a complete word ends here.
type trieNode struct {
	children map[rune]*trieNode
	isEnd    bool
}

func newTrieNode() *trieNode {
	return &trieNode{children: make(map[rune]*trieNode)}
}

// trie is a prefix tree holding command names.
type trie struct {
	root *trieNode
}

func newTrie() *trie {
	return &trie{root: newTrieNode()}
}

// Insert adds a word to the trie, creating intermediate nodes as needed.
func (t *trie) Insert(word string) {
	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			node.children[ch] = newTrieNode()
		}
		node = node.children[ch]
	}
	node.isEnd = true
}

// FindByPrefix returns all words in the trie that start with prefix, sorted.
func (t *trie) FindByPrefix(prefix string) []string {
	node := t.root
	for _, ch := range prefix {
		child, ok := node.children[ch]
		if !ok {
			return nil
		}
		node = child
	}
	var results []string
	node.collect(prefix, &results)
	sort.Strings(results)
	return results
}

// collect recursively gathers all complete words under this node.
func (n *trieNode) collect(prefix string, results *[]string) {
	if n.isEnd {
		*results = append(*results, prefix)
	}
	for ch, child := range n.children {
		child.collect(prefix+string(ch), results)
	}
}
