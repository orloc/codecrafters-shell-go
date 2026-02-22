package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

var commandTrie *trie

func initCommandTrie() {
	commandTrie = newTrie()
	for name := range registry {
		commandTrie.Insert(name)
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			commandTrie.Insert(e.Name())
		}
	}
}

type builtinCompleter struct {
	lastPrefix string
	tabCount   int
}

func (b *builtinCompleter) Do(line []rune, pos int) ([][]rune, int) {
	prefix := string(line[:pos])
	if strings.Contains(prefix, " ") {
		return nil, 0
	}

	matches := commandTrie.FindByPrefix(prefix)

	if len(matches) == 0 {
		fmt.Fprint(os.Stderr, "\x07")
		b.lastPrefix = ""
		b.tabCount = 0
		return nil, 0
	}

	if len(matches) == 1 {
		b.lastPrefix = ""
		b.tabCount = 0
		suffix := matches[0][len(prefix):] + " "
		return [][]rune{[]rune(suffix)}, len(prefix)
	}

	// Multiple matches — compute longest common prefix
	lcp := matches[0]
	for _, m := range matches[1:] {
		for !strings.HasPrefix(m, lcp) {
			lcp = lcp[:len(lcp)-1]
		}
	}

	if len(lcp) > len(prefix) {
		// Complete to LCP
		b.lastPrefix = ""
		b.tabCount = 0
		suffix := lcp[len(prefix):]
		// Check if completing to the LCP leaves a single match
		if lcp == matches[0] && len(matches) == 1 {
			suffix += " "
		}
		return [][]rune{[]rune(suffix)}, len(prefix)
	}

	// LCP equals prefix — nothing to complete, use double-TAB listing
	if prefix == b.lastPrefix {
		b.tabCount++
	} else {
		b.tabCount = 1
	}
	b.lastPrefix = prefix

	if b.tabCount == 1 {
		fmt.Fprint(os.Stderr, "\x07")
		return nil, 0
	}

	// Second consecutive TAB: list all matches
	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(matches, "  "), prefix)
	b.lastPrefix = ""
	b.tabCount = 0
	return nil, 0
}
