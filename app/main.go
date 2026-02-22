package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
			continue
		}

		// first see if the command we got exists on the file system
		// if it does and its executable - we should run it with the args passed to us
		cmd := exec.Command(name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				fmt.Printf("%s: command not found\n", name)
				continue
			}
			fmt.Println(err)
			continue
		}
	}
}

func trimInput(s string) (string, []string) {
	s = strings.TrimSpace(s)
	var (
		args       []string
		current    strings.Builder
		inSingleQ  bool
		inDoubleQ  bool
		hasContent = false
	)

	for i := 0; i < len(s); i++ {
		ch := s[i]

		switch {
		case ch == '\'' && !inDoubleQ && inSingleQ:
			inSingleQ = false
		case ch == '\'' && !inDoubleQ:
			inSingleQ = true
			hasContent = true
		case ch == '"' && !inSingleQ && inDoubleQ:
			inDoubleQ = false
		case ch == '"' && !inSingleQ:
			inDoubleQ = true
			hasContent = true
		case ch == ' ' && !inSingleQ && !inDoubleQ:
			if hasContent {
				args = append(args, current.String())
				current.Reset()
				hasContent = false
			}
		default:
			current.WriteByte(ch)
			hasContent = true
		}
	}

	if hasContent {
		args = append(args, current.String())
	}

	if len(args) == 0 {
		return "", nil
	}

	return args[0], args[1:]
}
