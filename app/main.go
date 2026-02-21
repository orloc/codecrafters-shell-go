package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		processCmd(input)
	}
}

func processCmd(s string) {
	// strip the \n
	sTrimed := strings.Replace(strings.TrimSpace(s), "\n", "", -1)
	switch sTrimed {
	case "exit":
		os.Exit(0)
	default:
		fmt.Printf("%s: command not found\n", sTrimed)
	}
}
