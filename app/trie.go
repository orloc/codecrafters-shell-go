package main

import "sort"

type trieNode struct {
	children map[rune]*trieNode
	isEnd    bool
}

func newTrieNode() *trieNode {
	return &trieNode{children: make(map[rune]*trieNode)}
}

type trie struct {
	root *trieNode
}

func newTrie() *trie {
	return &trie{root: newTrieNode()}
}

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

func (n *trieNode) collect(prefix string, results *[]string) {
	if n.isEnd {
		*results = append(*results, prefix)
	}
	for ch, child := range n.children {
		child.collect(prefix+string(ch), results)
	}
}
