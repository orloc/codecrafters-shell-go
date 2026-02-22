// completer.go — TAB completion for the shell prompt.
//
// On startup, initCommandTrie populates a prefix trie with all builtin
// command names and external executables found on PATH (directories are
// scanned concurrently via goroutines).
//
// builtinCompleter implements the readline.AutoCompleter interface:
//
//	single match   → complete the word + trailing space
//	multiple matches, LCP longer than prefix → complete to LCP
//	multiple matches, LCP == prefix:
//	    first TAB  → bell
//	    second TAB → list all matches
//	no matches     → bell
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var commandTrie *trie

// initCommandTrie builds the trie used for TAB completion. Builtins are
// inserted first, then PATH directories are scanned concurrently. A single
// goroutine drains the channel and inserts names into the trie (not
// goroutine-safe) to avoid locking.
func initCommandTrie() {
	commandTrie = newTrie()

	// Builtins go in first so they're always completable.
	for name := range registry {
		commandTrie.Insert(name)
	}

	// Scan PATH directories in parallel; feed names through a channel.
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

	// Single-threaded insert — trie is not goroutine-safe.
	for name := range names {
		commandTrie.Insert(name)
	}
}

// builtinCompleter implements readline.AutoCompleter. It tracks consecutive
// TAB presses to distinguish "complete" from "list all matches".
type builtinCompleter struct {
	lastPrefix string
	tabCount   int
}

// Do is called by readline on each TAB press. It returns candidate suffixes
// to append and the length of the prefix to replace.
func (b *builtinCompleter) Do(line []rune, pos int) ([][]rune, int) {
	prefix := string(line[:pos])

	// Only complete the first word (command name).
	if strings.Contains(prefix, " ") {
		return nil, 0
	}

	matches := commandTrie.FindByPrefix(prefix)

	// No matches — ring the bell.
	if len(matches) == 0 {
		fmt.Fprint(os.Stderr, "\x07")
		b.lastPrefix = ""
		b.tabCount = 0
		return nil, 0
	}

	// Exact single match — complete with trailing space.
	if len(matches) == 1 {
		b.lastPrefix = ""
		b.tabCount = 0
		suffix := matches[0][len(prefix):] + " "
		return [][]rune{[]rune(suffix)}, len(prefix)
	}

	// Multiple matches — compute longest common prefix (LCP).
	lcp := matches[0]
	for _, m := range matches[1:] {
		for !strings.HasPrefix(m, lcp) {
			lcp = lcp[:len(lcp)-1]
		}
	}

	// LCP is longer than what's typed — complete to the LCP.
	if len(lcp) > len(prefix) {
		b.lastPrefix = ""
		b.tabCount = 0
		suffix := lcp[len(prefix):]
		if lcp == matches[0] && len(matches) == 1 {
			suffix += " "
		}
		return [][]rune{[]rune(suffix)}, len(prefix)
	}

	// LCP equals prefix — nothing new to complete; use double-TAB listing.
	if prefix == b.lastPrefix {
		b.tabCount++
	} else {
		b.tabCount = 1
	}
	b.lastPrefix = prefix

	// First TAB with ambiguity — bell.
	if b.tabCount == 1 {
		fmt.Fprint(os.Stderr, "\x07")
		return nil, 0
	}

	// Second consecutive TAB — list all matches below the prompt.
	fmt.Fprintf(os.Stdout, "\n%s\n$ %s", strings.Join(matches, "  "), prefix)
	b.lastPrefix = ""
	b.tabCount = 0
	return nil, 0
}
