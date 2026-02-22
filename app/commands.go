package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Command represents a builtin shell command.
type Command struct {
	Name string
	Run  func(args []string)
}

var registry map[string]Command

func init() {
	registry = map[string]Command{
		"cd": {
			Name: "cd",
			Run: func(args []string) {
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Println(err)
				}
				if len(args) == 0 {
					os.Chdir(home)
					return
				}

				path := args[0]

				if path == "~" {
					os.Chdir(home)
					return
				}

				if err = os.Chdir(path); err != nil {
					fmt.Printf("cd: %s: No such file or directory\n", path)
				}
			},
		},
		"pwd": {
			Name: "pwd",
			Run: func(args []string) {
				dir, err := os.Getwd()
				if err != nil {
					fmt.Println(err)
					return
				}

				fmt.Println(dir)
			},
		},
		"echo": {
			Name: "echo",
			Run: func(args []string) {
				fmt.Printf("%s\n", strings.Join(args, " "))
			},
		},
		"exit": {
			Name: "exit",
			Run: func(args []string) {
				os.Exit(0)
			},
		},
		"type": {
			Name: "type",
			Run: func(args []string) {
				arg := strings.Join(args, " ")
				if _, ok := registry[arg]; ok {
					fmt.Printf("%s is a shell builtin\n", arg)
					return
				}
				p, err := exec.LookPath(arg)
				if err != nil {
					fmt.Printf("%s: not found\n", arg)
					return
				}
				fmt.Printf("%s is %s\n", arg, p)
			},
		},
		"history": {
			Name: "history",
			Run: func(args []string) {
				n := 0
				if len(args) > 0 {
					fmt.Sscan(args[0], &n)
				}
				printHistory(n)
			},
		},
	}
}

// GetCommand looks up a builtin command by name.
func GetCommand(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}
