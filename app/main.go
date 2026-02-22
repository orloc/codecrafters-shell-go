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
		name, args := trimInput(input)
		if cmd, ok := GetCommand(name); ok {
			cmd.Run(args)
		} else {
			fmt.Printf("%s: command not found\n", name)
		}
	}
}

func trimInput(s string) (string, []string) {
	sTrimed := strings.Replace(strings.TrimSpace(s), "\n", "", -1)
	// split the cmd arg out from params
	args := strings.Split(sTrimed, " ")

	return args[0], args[1:]
}
