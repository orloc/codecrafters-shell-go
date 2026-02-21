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
		processCmd(trimInput(input))
	}
}

func trimInput(s string) (string, []string) {
	sTrimed := strings.Replace(strings.TrimSpace(s), "\n", "", -1)
	// split the cmd arg out from params
	args := strings.Split(sTrimed, " ")

	return sTrimed, args
}

func processCmd(s string, args []string) {
	// strip the \n
	switch args[0] {
	case "echo":
		fmt.Printf("%s\n", strings.Join(args[1:], " "))
	case "exit":
		os.Exit(0)
	default:
		fmt.Printf("%s: command not found\n", s)
	}
}
