package main

import (
	"fmt"
	"strings"
)

// --- Types ---

// parsedCommand holds the result of parsing a single command segment.
type parsedCommand struct {
	Name      string
	Args      []string
	Redirects []Redirect
}

// --- Entry points (called from handleInput / startSegment) ---

// parseCommand parses a raw input segment into a command name, arguments,
// and any I/O redirections.
func parseCommand(input string) (parsedCommand, error) {
	cmdPart, redirects, err := parseRedirection(input)
	if err != nil {
		return parsedCommand{}, err
	}
	name, args := trimInput(cmdPart)
	return parsedCommand{Name: name, Args: args, Redirects: redirects}, nil
}

// parsePipeline splits the input on unquoted, unescaped '|' characters,
// respecting single quotes, double quotes, and backslash escapes.
func parsePipeline(s string) []string {
	var (
		q        quoteTracker
		buf      strings.Builder
		segments []string
	)
	for i := 0; i < len(s); i++ {
		if newI, ok := q.skipEscape(s, i, &buf); ok {
			i = newI
			continue
		}
		if q.toggleQuote(s[i], &buf) {
			continue
		}
		if s[i] == '|' && !q.IsQuoted() {
			segments = append(segments, buf.String())
			buf.Reset()
			continue
		}
		buf.WriteByte(s[i])
	}
	segments = append(segments, buf.String())
	return segments
}

// --- Mid-level parsing ---

// parseRedirection separates redirect operators (>, >>, 1>, 2>, etc.) from
// the command text, returning the command portion and a slice of Redirects.
func parseRedirection(s string) (string, []Redirect, error) {
	var (
		q         quoteTracker
		cmdPart   strings.Builder
		redirects []Redirect
	)

	for i := 0; i < len(s); i++ {
		if newI, ok := q.skipEscape(s, i, &cmdPart); ok {
			i = newI
			continue
		}
		if q.toggleQuote(s[i], &cmdPart) {
			continue
		}

		// Check for redirect operator (only outside quotes)
		if s[i] == '>' && !q.IsQuoted() {
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

		cmdPart.WriteByte(s[i])
	}

	if len(redirects) == 0 {
		return s, nil, nil
	}

	return cmdPart.String(), redirects, nil
}

// trimInput splits a command string into the command name and its arguments,
// resolving quotes and escapes via nextToken.
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

// --- Low-level scanning helpers ---

// quoteTracker tracks single/double quote state while scanning a shell string.
// Used by parsePipeline and parseRedirection to share quoting logic.
type quoteTracker struct {
	inSingle bool
	inDouble bool
}

func (q *quoteTracker) IsQuoted() bool {
	return q.inSingle || q.inDouble
}

// skipEscape handles a backslash at s[i] outside single quotes. It writes the
// raw backslash and escaped character to buf and returns the updated index.
// Returns false if s[i] is not an escape sequence.
func (q *quoteTracker) skipEscape(s string, i int, buf *strings.Builder) (int, bool) {
	if s[i] != '\\' || q.inSingle {
		return i, false
	}
	buf.WriteByte(s[i])
	if i+1 < len(s) {
		i++
		buf.WriteByte(s[i])
	}
	return i, true
}

// toggleQuote handles quote characters. If ch is a quote that should toggle
// state, it writes the char to buf and returns true.
func (q *quoteTracker) toggleQuote(ch byte, buf *strings.Builder) bool {
	if ch == '\'' && !q.inDouble {
		q.inSingle = !q.inSingle
		buf.WriteByte(ch)
		return true
	}
	if ch == '"' && !q.inSingle {
		q.inDouble = !q.inDouble
		buf.WriteByte(ch)
		return true
	}
	return false
}

// nextToken parses one shell token starting at s[pos], resolving single quotes,
// double quotes, and backslash escapes. Returns the resolved value and the
// position where scanning stopped. Stops at unquoted space, newline, or any
// unquoted byte in stops.
//
// NOTE: nextToken resolves quotes/escapes (strips quote chars, interprets
// backslash sequences) unlike quoteTracker which preserves them raw. The two
// serve different purposes so nextToken keeps its own state.
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
