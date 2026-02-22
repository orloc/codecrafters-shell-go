package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var commandTrie *trie

func initCommandTrie() {
	commandTrie = newTrie()
	for name := range registry {
		commandTrie.Insert(name)
	}

	dirs := filepath.SplitList(os.Getenv("PATH"))
	names := make(chan string, 64)

	var wg sync.WaitGroup
	for _, dir := range dirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			entries, err := os.ReadDir(dir)
			if err != nil {
				return
			}
			for _, e := range entries {
				if !e.IsDir() {
					names <- e.Name()
				}
			}
		}(dir)
	}

	go func() {
		wg.Wait()
		close(names)
	}()

	for name := range names {
		commandTrie.Insert(name)
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
