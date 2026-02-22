package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var validCmds = map[string]struct{}{
	"type": {},
	"exit": {},
	"echo": {},
}

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

	return args[0], args[1:]
}

func processCmd(s string, args []string) {
	// strip the \n
	switch s {
	case "type":
		arg := strings.Join(args, " ")
		if _, ok := validCmds[arg]; !ok {
			fmt.Printf("%s: not found\n", arg)
		} else {
			fmt.Printf("%s is a shell builtin\n", arg)
		}
	case "echo":
		fmt.Printf("%s\n", strings.Join(args, " "))
	case "exit":
		os.Exit(0)
	default:
		fmt.Printf("%s: command not found\n", s)
	}
}
