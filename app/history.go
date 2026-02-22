package main

import "fmt"

var commandHistory []string

// recordHistory appends a raw input line to the history.
func recordHistory(input string) {
	commandHistory = append(commandHistory, input)
}

// printHistory prints all history entries in the standard numbered format.
func printHistory() {
	for i, cmd := range commandHistory {
		fmt.Printf("%5d  %s\n", i+1, cmd)
	}
}
