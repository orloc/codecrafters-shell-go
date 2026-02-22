package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/chzyer/readline"
)

func main() {
	initCommandTrie()
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: &builtinCompleter{},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		handleInput(line)
	}
}

func handleInput(input string) {
	cmdPart, redirects, err := parseRedirection(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	name, args := trimInput(cmdPart)

	stdout, stderr, cleanup, err := openRedirects(redirects)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer cleanup()

	if cmd, ok := GetCommand(name); ok {
		os.Stdout = stdout
		os.Stderr = stderr
		cmd.Run(args)
		return
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err = cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Printf("%s: command not found\n", name)
		}
	}
}
