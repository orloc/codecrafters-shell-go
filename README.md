# gosh - A POSIX Shell in Go

A shell implementation built from scratch in Go as part of the [CodeCrafters "Build Your Own Shell"](https://app.codecrafters.io/courses/shell/overview) challenge. Supports command execution, pipelines, I/O redirection, quoting/escaping, TAB completion, and persistent command history.

## Features

- **Builtin commands**: `cd`, `pwd`, `echo`, `exit`, `type`, `history`
- **External commands**: PATH lookup and execution via `os/exec`
- **Pipelines**: `cmd1 | cmd2 | cmd3` with arbitrary depth
- **I/O redirection**: `>`, `>>`, `1>`, `2>`, `1>>`, `2>>`
- **Quoting**: single quotes, double quotes, backslash escapes (POSIX-compliant)
- **TAB completion**: prefix trie with single-TAB complete, double-TAB listing, LCP completion
- **Command history**: in-memory tracking with file persistence (`HISTFILE`), `history -r/-w/-a`
- **Signal handling**: graceful history save on SIGTERM/SIGHUP

## Architecture

```
main.go          entry point, readline loop, signal handling
  -> handleInput
       -> parsePipeline     split on unquoted '|'        (parser.go)
       -> executePipeline   pipe execution via os.Pipe    (pipeline.go)
       -> parseCommand      redirections + tokenization   (parser.go)
       -> openRedirects     file-based I/O redirection    (redirect.go)
       -> GetCommand        builtin lookup                (commands.go)
       -> exec.Command      external process fallback

completer.go     TAB completion (readline.AutoCompleter)
trie.go          prefix trie for command name lookup
history.go       History struct with file I/O (read/write/append)
```

### File overview

| File | Lines | Purpose |
|------|------:|---------|
| `parser.go` | 267 | Pipeline splitting, redirection parsing, tokenization, quote handling |
| `pipeline.go` | 205 | Multi-segment pipe execution with goroutines for builtins |
| `completer.go` | 144 | TAB completion with concurrent PATH scanning |
| `history.go` | 122 | In-memory history with file persistence and flush tracking |
| `commands.go` | 119 | Builtin command registry |
| `main.go` | 118 | Entry point, readline loop, HISTFILE/signal handling |
| `trie.go` | 66 | Prefix trie data structure |
| `redirect.go` | 56 | I/O redirection file management |

### Key design decisions

- **Single `package main`**: flat structure, one concern per file. No internal packages — this is an application, not a library.
- **Non-blocking pipelines**: external commands use `cmd.Start()`, builtins run in goroutines with swapped `os.Stdout`.
- **Two quote-handling modes**: `quoteTracker` preserves raw quote chars (for pipeline/redirection splitting), `nextToken` resolves and strips them (for tokenization).
- **History flush tracking**: `lastFlushed` index ensures `AppendFile` only writes new entries, preventing duplicates across multiple appends.
- **Concurrent PATH scanning**: goroutines scan PATH directories in parallel, feeding a channel that a single goroutine drains into the trie (not goroutine-safe).

## Usage

```sh
go build -o gosh ./app/
./gosh
```

With persistent history:

```sh
HISTFILE=~/.gosh_history ./gosh
```

### History commands

```sh
history          # print all history
history 10       # print last 10 entries
history -r file  # read file into memory
history -w file  # write all history to file
history -a file  # append new entries to file
```

## Testing

```sh
go test ./app/ -v
```

134 tests covering parsing, pipelines, redirection, completion, history, and builtins.

## Requirements

- Go 1.22+
- [github.com/chzyer/readline](https://github.com/chzyer/readline) (interactive line editing)

## License

MIT License. See [LICENSE](LICENSE) for details.
