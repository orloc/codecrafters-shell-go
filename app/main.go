package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
		handleInput(input)
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

// openRedirects opens files for each redirect and returns the stdout/stderr
// writers to use. The returned cleanup function closes all opened files and
// restores os.Stdout/os.Stderr to their original values.
func openRedirects(redirects []Redirect) (stdout, stderr *os.File, cleanup func(), err error) {
	origStdout := os.Stdout
	origStderr := os.Stderr
	stdout = origStdout
	stderr = origStderr

	var files []*os.File
	for _, r := range redirects {
		var f *os.File
		if r.Op == ">" {
			f, err = os.Create(r.File)
		} else {
			f, err = os.OpenFile(r.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		}
		if err != nil {
			for _, cf := range files {
				cf.Close()
			}
			return nil, nil, nil, err
		}
		files = append(files, f)
		if r.Fd == 1 {
			stdout = f
		} else {
			stderr = f
		}
	}

	cleanup = func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		for _, f := range files {
			f.Close()
		}
	}

	return stdout, stderr, cleanup, nil
}
