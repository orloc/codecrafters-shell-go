// history.go — in-memory command history with file persistence.
//
// History tracks entries as a simple []string. File operations:
//   - ReadFile:   load from disk, appending to in-memory list
//   - WriteFile:  overwrite file with all entries
//   - AppendFile: append only new (unflushed) entries since last call
//
// The lastFlushed index tracks the boundary for AppendFile so repeated
// calls don't duplicate entries.
package main

import (
	"bufio"
	"fmt"
	"os"
)

// History tracks shell command history in memory with file I/O support.
type History struct {
	entries     []string
	lastFlushed int
}

// NewHistory creates an empty history.
func NewHistory() *History {
	return &History{}
}

// Record appends a raw input line to the history.
func (h *History) Record(input string) {
	h.entries = append(h.entries, input)
}

// ReadFile reads lines from path and appends them to the in-memory
// history. Empty lines are skipped.
func (h *History) ReadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}
	return scanner.Err()
}

// WriteFile writes all in-memory history entries to path,
// creating the file if it doesn't exist. Each entry is followed by a newline.
func (h *History) WriteFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range h.entries {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// AppendFile appends only new (unflushed) in-memory history entries
// to path, creating the file if it doesn't exist.
func (h *History) AppendFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range h.entries[h.lastFlushed:] {
		fmt.Fprintln(w, line)
	}
	h.lastFlushed = len(h.entries)
	return w.Flush()
}

// Print prints the last n history entries (or all if n <= 0).
func (h *History) Print(n int) {
	start := 0
	if n > 0 && n < len(h.entries) {
		start = len(h.entries) - n
	}
	for i := start; i < len(h.entries); i++ {
		fmt.Printf("%5d  %s\n", i+1, h.entries[i])
	}
}
