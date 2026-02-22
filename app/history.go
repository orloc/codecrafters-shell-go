package main

import "fmt"

var commandHistory []string

// recordHistory appends a raw input line to the history.
func recordHistory(input string) {
	commandHistory = append(commandHistory, input)
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
