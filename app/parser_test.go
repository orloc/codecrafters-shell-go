package main

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantName      string
		wantArgs      []string
		wantRedirects []Redirect
		wantErr       bool
	}{
		{
			name:     "simple command",
			input:    "echo hello",
			wantName: "echo",
			wantArgs: []string{"hello"},
		},
		{
			name:          "command with redirect",
			input:         "echo hello > file.txt",
			wantName:      "echo",
			wantArgs:      []string{"hello"},
			wantRedirects: []Redirect{{Fd: 1, Op: ">", File: "file.txt"}},
		},
		{
			name:     "command with quoted args and redirect",
			input:    `echo "hello world" > file.txt`,
			wantName: "echo",
			wantArgs: []string{"hello world"},
			wantRedirects: []Redirect{{Fd: 1, Op: ">", File: "file.txt"}},
		},
		{
			name:    "missing redirect target is error",
			input:   "echo >",
			wantErr: true,
		},
		{
			name:  "whitespace only",
			input: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.Name != tt.wantName {
				t.Errorf("parseCommand() Name = %q, want %q", got.Name, tt.wantName)
			}
			if len(got.Args) != len(tt.wantArgs) {
				t.Fatalf("parseCommand() Args len = %d, want %d\n  got:  %q\n  want: %q", len(got.Args), len(tt.wantArgs), got.Args, tt.wantArgs)
			}
			for i := range got.Args {
				if got.Args[i] != tt.wantArgs[i] {
					t.Errorf("parseCommand() Args[%d] = %q, want %q", i, got.Args[i], tt.wantArgs[i])
				}
			}
			if len(got.Redirects) != len(tt.wantRedirects) {
				t.Fatalf("parseCommand() Redirects len = %d, want %d", len(got.Redirects), len(tt.wantRedirects))
			}
			for i, r := range got.Redirects {
				want := tt.wantRedirects[i]
				if r.Fd != want.Fd || r.Op != want.Op || r.File != want.File {
					t.Errorf("parseCommand() Redirects[%d] = %+v, want %+v", i, r, want)
				}
			}
		})
	}
}

func TestParsePipeline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "no pipe",
			input: "echo hello",
			want:  []string{"echo hello"},
		},
		{
			name:  "single pipe",
			input: "cat file | wc",
			want:  []string{"cat file ", " wc"},
		},
		{
			name:  "pipe inside double quotes",
			input: `echo "a|b"`,
			want:  []string{`echo "a|b"`},
		},
		{
			name:  "pipe inside single quotes",
			input: "echo 'a|b'",
			want:  []string{"echo 'a|b'"},
		},
		{
			name:  "escaped pipe",
			input: `echo a\|b`,
			want:  []string{`echo a\|b`},
		},
		{
			name:  "multiple pipes",
			input: "a | b | c",
			want:  []string{"a ", " b ", " c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePipeline(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parsePipeline(%q) returned %d segments, want %d\n  got:  %q\n  want: %q", tt.input, len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parsePipeline(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTrimInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:    "simple command no args",
			input:   "echo\n",
			wantCmd: "echo",
		},
		{
			name:     "command with one arg",
			input:    "echo hello\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello"},
		},
		{
			name:     "command with multiple args",
			input:    "echo hello world\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello", "world"},
		},
		{
			name:     "command with leading/trailing whitespace",
			input:    "  type echo  \n",
			wantCmd:  "type",
			wantArgs: []string{"echo"},
		},
		{
			name:    "exit command",
			input:   "exit\n",
			wantCmd: "exit",
		},
		{
			name:     "single-quoted arg preserves spaces",
			input:    "echo 'hello world'\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello world"},
		},
		{
			name:     "multiple single-quoted args",
			input:    "echo 'hello' 'world'\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello", "world"},
		},
		{
			name:     "mixed quoted and unquoted args",
			input:    "echo hello 'big world' foo\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello", "big world", "foo"},
		},
		{
			name:     "empty single-quoted string is dropped",
			input:    "echo '' foo\n",
			wantCmd:  "echo",
			wantArgs: []string{"foo"},
		},
		{
			name:     "adjacent quotes concatenate",
			input:    "echo 'hello''world'\n",
			wantCmd:  "echo",
			wantArgs: []string{"helloworld"},
		},
		{
			name:     "multiple spaces between args",
			input:    "echo   hello   world\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello", "world"},
		},
		{
			name:     "double-quoted arg preserves spaces",
			input:    "echo \"hello    world\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello    world"},
		},
		{
			name:     "double-quoted strings concatenate",
			input:    "echo \"hello\"\"world\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"helloworld"},
		},
		{
			name:     "double-quoted separate args",
			input:    "echo \"hello\" \"world\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello", "world"},
		},
		{
			name:     "single quote inside double quotes is literal",
			input:    "echo \"shell's test\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"shell's test"},
		},
		{
			name:     "double quote inside single quotes is literal",
			input:    "echo 'say \"hi\"'\n",
			wantCmd:  "echo",
			wantArgs: []string{"say \"hi\""},
		},
		{
			name:     "empty double-quoted string is dropped",
			input:    "echo \"\" foo\n",
			wantCmd:  "echo",
			wantArgs: []string{"foo"},
		},
		{
			name:     "double-quoted preserves tabs",
			input:    "echo \"hello\tworld\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello\tworld"},
		},
		{
			name:     "mixed single and double quotes",
			input:    "echo 'hello' \"world\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello", "world"},
		},
		{
			name:     "backslash escapes space to join args",
			input:    "echo three\\ \\ \\ spaces\n",
			wantCmd:  "echo",
			wantArgs: []string{"three   spaces"},
		},
		{
			name:     "backslash preserves first space only",
			input:    "echo before\\     after\n",
			wantCmd:  "echo",
			wantArgs: []string{"before ", "after"},
		},
		{
			name:     "backslash letter is literal letter",
			input:    "echo test\\nexample\n",
			wantCmd:  "echo",
			wantArgs: []string{"testnexample"},
		},
		{
			name:     "backslash escapes backslash",
			input:    "echo hello\\\\world\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello\\world"},
		},
		{
			name:     "backslash escapes single quotes",
			input:    "echo \\'hello\\'\n",
			wantCmd:  "echo",
			wantArgs: []string{"'hello'"},
		},
		{
			name:     "backslash escapes double quotes",
			input:    "echo \\\"hello\\\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"\"hello\""},
		},
		{
			name:     "backslash in double quotes escapes backslash",
			input:    "echo \"hello\\\\world\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello\\world"},
		},
		{
			name:     "backslash in double quotes escapes double quote",
			input:    "echo \"hello\\\"world\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"hello\"world"},
		},
		{
			name:     "backslash in double quotes is literal for other chars",
			input:    "echo \"test\\nvalue\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"test\\nvalue"},
		},
		{
			name:     "backslash in double quotes literal for dollar",
			input:    "echo \"price\\$5\"\n",
			wantCmd:  "echo",
			wantArgs: []string{"price\\$5"},
		},
		{
			name:  "empty input",
			input: "   \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := trimInput(tt.input)
			if gotCmd != tt.wantCmd {
				t.Errorf("trimInput() cmd = %q, want %q", gotCmd, tt.wantCmd)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("trimInput() args len = %d, want %d", len(gotArgs), len(tt.wantArgs))
				return
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("trimInput() args[%d] = %q, want %q", i, gotArgs[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestParseRedirection(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantCmd       string
		wantRedirects []Redirect
		wantErr       bool
	}{
		{
			name:          "stdout truncate",
			input:         "echo hello > file.txt",
			wantCmd:       "echo hello ",
			wantRedirects: []Redirect{{Fd: 1, Op: ">", File: "file.txt"}},
		},
		{
			name:          "stdout append",
			input:         "echo hello >> file.txt",
			wantCmd:       "echo hello ",
			wantRedirects: []Redirect{{Fd: 1, Op: ">>", File: "file.txt"}},
		},
		{
			name:          "explicit stdout truncate 1>",
			input:         "echo hello 1> file.txt",
			wantCmd:       "echo hello ",
			wantRedirects: []Redirect{{Fd: 1, Op: ">", File: "file.txt"}},
		},
		{
			name:          "explicit stdout append 1>>",
			input:         "echo hello 1>> file.txt",
			wantCmd:       "echo hello ",
			wantRedirects: []Redirect{{Fd: 1, Op: ">>", File: "file.txt"}},
		},
		{
			name:          "stderr truncate 2>",
			input:         "cmd 2> err.txt",
			wantCmd:       "cmd ",
			wantRedirects: []Redirect{{Fd: 2, Op: ">", File: "err.txt"}},
		},
		{
			name:          "stderr append 2>>",
			input:         "cmd 2>> err.txt",
			wantCmd:       "cmd ",
			wantRedirects: []Redirect{{Fd: 2, Op: ">>", File: "err.txt"}},
		},
		{
			name:    "multiple redirects stdout and stderr",
			input:   "cmd > out.txt 2> err.txt",
			wantCmd: "cmd  ",
			wantRedirects: []Redirect{
				{Fd: 1, Op: ">", File: "out.txt"},
				{Fd: 2, Op: ">", File: "err.txt"},
			},
		},
		{
			name:    "redirect inside double quotes is literal",
			input:   `echo "hello > world"`,
			wantCmd: `echo "hello > world"`,
		},
		{
			name:    "redirect inside single quotes is literal",
			input:   `echo 'hello > world'`,
			wantCmd: `echo 'hello > world'`,
		},
		{
			name:    "missing file path after redirect",
			input:   "echo hello >",
			wantErr: true,
		},
		{
			name:    "escaped redirect is literal",
			input:   `echo hello \> world`,
			wantCmd: `echo hello \> world`,
		},
		{
			name:    "no redirect",
			input:   "echo hello world",
			wantCmd: "echo hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdPart, redirects, err := parseRedirection(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRedirection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if cmdPart != tt.wantCmd {
				t.Errorf("parseRedirection() cmdPart = %q, want %q", cmdPart, tt.wantCmd)
			}
			if len(redirects) != len(tt.wantRedirects) {
				t.Errorf("parseRedirection() redirects len = %d, want %d\n  got:  %+v\n  want: %+v", len(redirects), len(tt.wantRedirects), redirects, tt.wantRedirects)
				return
			}
			for i, r := range redirects {
				want := tt.wantRedirects[i]
				if r.Fd != want.Fd || r.Op != want.Op || r.File != want.File {
					t.Errorf("parseRedirection() redirects[%d] = %+v, want %+v", i, r, want)
				}
			}
		})
	}
}

func TestNextToken(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pos     int
		stops   string
		wantTok string
		wantPos int
	}{
		{
			name:    "simple word",
			input:   "hello world",
			pos:     0,
			wantTok: "hello",
			wantPos: 5,
		},
		{
			name:    "start at offset",
			input:   "hello world",
			pos:     6,
			wantTok: "world",
			wantPos: 11,
		},
		{
			name:    "single-quoted string",
			input:   "'hello world' rest",
			pos:     0,
			wantTok: "hello world",
			wantPos: 13,
		},
		{
			name:    "double-quoted string",
			input:   `"hello world" rest`,
			pos:     0,
			wantTok: "hello world",
			wantPos: 13,
		},
		{
			name:    "backslash escape outside quotes",
			input:   `hello\ world rest`,
			pos:     0,
			wantTok: "hello world",
			wantPos: 12,
		},
		{
			name:    "backslash in double quotes escapes quote",
			input:   `"say \"hi\""`,
			pos:     0,
			wantTok: `say "hi"`,
			wantPos: 12,
		},
		{
			name:    "backslash in double quotes literal for normal char",
			input:   `"test\nval"`,
			pos:     0,
			wantTok: `test\nval`,
			wantPos: 11,
		},
		{
			name:    "stops at stop character",
			input:   "file.txt>rest",
			pos:     0,
			stops:   ">",
			wantTok: "file.txt",
			wantPos: 8,
		},
		{
			name:    "stops at newline",
			input:   "hello\nworld",
			pos:     0,
			wantTok: "hello",
			wantPos: 5,
		},
		{
			name:    "empty input from pos",
			input:   "hello",
			pos:     5,
			wantTok: "",
			wantPos: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTok, gotPos := nextToken(tt.input, tt.pos, tt.stops)
			if gotTok != tt.wantTok {
				t.Errorf("nextToken() token = %q, want %q", gotTok, tt.wantTok)
			}
			if gotPos != tt.wantPos {
				t.Errorf("nextToken() pos = %d, want %d", gotPos, tt.wantPos)
			}
		})
	}
}
