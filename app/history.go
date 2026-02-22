package main

import (
	"bufio"
	"fmt"
	"os"
)

var commandHistory []string

// recordHistory appends a raw input line to the history.
func recordHistory(input string) {
	commandHistory = append(commandHistory, input)
}

// readHistoryFile reads lines from path and appends them to the in-memory
// history. Empty lines are skipped.
func readHistoryFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			commandHistory = append(commandHistory, line)
		}
	}
	return scanner.Err()
}

// writeHistoryFile writes all in-memory history entries to path,
// creating the file if it doesn't exist. Each entry is followed by a newline.
func writeHistoryFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range commandHistory {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// appendHistoryFile appends all in-memory history entries to path,
// creating the file if it doesn't exist.
func appendHistoryFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range commandHistory {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

// printHistory prints the last n history entries (or all if n <= 0).
func printHistory(n int) {
	start := 0
	if n > 0 && n < len(commandHistory) {
		start = len(commandHistory) - n
	}
	for i := start; i < len(commandHistory); i++ {
		fmt.Printf("%5d  %s\n", i+1, commandHistory[i])
	}
}
