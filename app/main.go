package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/chzyer/readline"
)

// main starts the shell: populates the command trie for TAB completion,
// sets up readline, and enters the read-eval loop.
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
		if err != nil { // EOF or ^C
			break
		}
		handleInput(line)
	}
}

// handleInput processes a single input line through the shell's execution
// pipeline:
//
//	input
//	  -> parsePipeline       split on unquoted '|'
//	     len > 1?  ----yes----> executePipeline (see pipeline.go)
//	     |
//	     no (single command)
//	     |
//	  -> parseCommand        parse redirections + tokenize into name/args
//	  -> openRedirects       open redirect target files for stdout/stderr
//	  -> GetCommand          look up builtin; if found, run in-process
//	     or exec.Command     otherwise spawn an external process
func handleInput(input string) {
	recordHistory(input)

	// Split on pipes first; delegate multi-segment pipelines.
	segments := parsePipeline(input)
	if len(segments) > 1 {
		executePipeline(segments)
		return
	}

	// Single command: parse into name, args, and redirections.
	parsed, err := parseCommand(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// Open redirect target files; cleanup restores original stdout/stderr.
	stdout, stderr, cleanup, err := openRedirects(parsed.Redirects)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer cleanup()

	// Try builtins first (cd, echo, pwd, type, exit).
	if cmd, ok := GetCommand(parsed.Name); ok {
		os.Stdout = stdout
		os.Stderr = stderr
		cmd.Run(parsed.Args)
		return
	}

	// Fall back to external command lookup via PATH.
	cmd := exec.Command(parsed.Name, parsed.Args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err = cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Printf("%s: command not found\n", parsed.Name)
		}
	}
}
