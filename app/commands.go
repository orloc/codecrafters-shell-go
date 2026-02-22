// commands.go — builtin command registry (cd, pwd, echo, exit, type, history).
//
// newRegistry() builds the map; GetCommand() looks up by name. Each command
// is a simple function value — no interface needed at this scale.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Command represents a builtin shell command.
type Command struct {
	Run func(args []string)
}

var registry map[string]Command

// newRegistry builds the builtin command registry.
func newRegistry() {
	registry = map[string]Command{
		"cd": {
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
			Run: func(args []string) {
				fmt.Printf("%s\n", strings.Join(args, " "))
			},
		},
		"exit": {
			Run: func(args []string) {
				saveHistory()
				os.Exit(0)
			},
		},
		"type": {
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
			Run: func(args []string) {
				if len(args) >= 2 {
					switch args[0] {
					case "-r":
						if err := hist.ReadFile(args[1]); err != nil {
							fmt.Fprintf(os.Stderr, "history: %s\n", err)
						}
						return
					case "-w":
						if err := hist.WriteFile(args[1]); err != nil {
							fmt.Fprintf(os.Stderr, "history: %s\n", err)
						}
						return
					case "-a":
						if err := hist.AppendFile(args[1]); err != nil {
							fmt.Fprintf(os.Stderr, "history: %s\n", err)
						}
						return
					}
				}
				n := 0
				if len(args) > 0 {
					fmt.Sscan(args[0], &n)
				}
				hist.Print(n)
			},
		},
	}
}

// GetCommand looks up a builtin command by name.
func GetCommand(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}
