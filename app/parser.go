package main

import (
	"fmt"
	"os"
	"strings"
)

type Redirect struct {
	Fd   int    // 1 = stdout, 2 = stderr
	Op   string // ">" or ">>"
	File string // target file path
}

// nextToken parses one shell token starting at s[pos], resolving single quotes,
// double quotes, and backslash escapes. Returns the resolved value and the
// position where scanning stopped. Stops at unquoted space, newline, or any
// unquoted byte in stops.
func nextToken(s string, pos int, stops string) (string, int) {
	var (
		buf       strings.Builder
		inSingleQ bool
		inDoubleQ bool
	)
	i := pos
	for i < len(s) {
		ch := s[i]

		if ch == '\n' {
			break
		}

		if ch == '\\' && !inSingleQ {
			if inDoubleQ {
				if i+1 < len(s) && (s[i+1] == '\\' || s[i+1] == '"') {
					i++
					buf.WriteByte(s[i])
				} else {
					buf.WriteByte(ch)
				}
			} else {
				if i+1 < len(s) {
					i++
					buf.WriteByte(s[i])
				}
			}
			i++
			continue
		}

		if ch == '\'' && !inDoubleQ {
			inSingleQ = !inSingleQ
			i++
			continue
		}

		if ch == '"' && !inSingleQ {
			inDoubleQ = !inDoubleQ
			i++
			continue
		}

		if !inSingleQ && !inDoubleQ {
			if ch == ' ' {
				break
			}
			if len(stops) > 0 && strings.IndexByte(stops, ch) >= 0 {
				break
			}
		}

		buf.WriteByte(ch)
		i++
	}
	return buf.String(), i
}

func parseRedirection(s string) (string, []Redirect, error) {
	var (
		cmdPart   strings.Builder
		redirects []Redirect
		inSingleQ bool
		inDoubleQ bool
	)

	for i := 0; i < len(s); i++ {
		ch := s[i]

		// Handle backslash escapes (not inside single quotes)
		if ch == '\\' && !inSingleQ {
			cmdPart.WriteByte(ch)
			if i+1 < len(s) {
				i++
				cmdPart.WriteByte(s[i])
			}
			continue
		}

		// Handle quotes
		if ch == '\'' && !inDoubleQ {
			inSingleQ = !inSingleQ
			cmdPart.WriteByte(ch)
			continue
		}
		if ch == '"' && !inSingleQ {
			inDoubleQ = !inDoubleQ
			cmdPart.WriteByte(ch)
			continue
		}

		// Check for redirect operator (only outside quotes)
		if ch == '>' && !inSingleQ && !inDoubleQ {
			fd := 1
			op := ">"

			// Check if preceded by fd digit (1 or 2)
			cmdStr := cmdPart.String()
			if len(cmdStr) > 0 && (cmdStr[len(cmdStr)-1] == '1' || cmdStr[len(cmdStr)-1] == '2') {
				fd = int(cmdStr[len(cmdStr)-1] - '0')
				cmdPart.Reset()
				cmdPart.WriteString(cmdStr[:len(cmdStr)-1])
			}

			// Check for >> (append mode)
			if i+1 < len(s) && s[i+1] == '>' {
				op = ">>"
				i++
			}

			// Skip spaces after operator
			i++
			for i < len(s) && s[i] == ' ' {
				i++
			}

			filePath, newPos := nextToken(s, i, ">")
			i = newPos - 1 // compensate for outer loop increment

			if filePath == "" {
				return "", nil, fmt.Errorf("syntax error near unexpected token 'newline'")
			}

			redirects = append(redirects, Redirect{
				Fd:   fd,
				Op:   op,
				File: filePath,
			})
			continue
		}

		cmdPart.WriteByte(ch)
	}

	if len(redirects) == 0 {
		return s, nil, nil
	}

	return cmdPart.String(), redirects, nil
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

func trimInput(s string) (string, []string) {
	s = strings.TrimSpace(s)
	var args []string
	i := 0
	for i < len(s) {
		for i < len(s) && s[i] == ' ' {
			i++
		}
		if i >= len(s) {
			break
		}
		token, newPos := nextToken(s, i, "")
		if token != "" {
			args = append(args, token)
		}
		i = newPos
	}
	if len(args) == 0 {
		return "", nil
	}
	return args[0], args[1:]
}
