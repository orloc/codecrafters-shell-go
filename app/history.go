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
