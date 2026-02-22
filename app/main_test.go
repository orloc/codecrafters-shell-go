package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

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
			name: "empty input",
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

// captureStdout runs fn and returns whatever it wrote to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestEchoCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "single word",
			args: []string{"hello"},
			want: "hello\n",
		},
		{
			name: "multiple words",
			args: []string{"hello", "world"},
			want: "hello world\n",
		},
		{
			name: "no args",
			args: []string{},
			want: "\n",
		},
	}

	cmd, ok := GetCommand("echo")
	if !ok {
		t.Fatal("echo command not found in registry")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				cmd.Run(tt.args)
			})
			if got != tt.want {
				t.Errorf("echo(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestTypeCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "builtin command echo",
			args: []string{"echo"},
			want: "echo is a shell builtin\n",
		},
		{
			name: "builtin command type",
			args: []string{"type"},
			want: "type is a shell builtin\n",
		},
		{
			name: "builtin command exit",
			args: []string{"exit"},
			want: "exit is a shell builtin\n",
		},
		{
			name: "unknown command",
			args: []string{"foo"},
			want: "foo: not found\n",
		},
	}

	cmd, ok := GetCommand("type")
	if !ok {
		t.Fatal("type command not found in registry")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				cmd.Run(tt.args)
			})
			if got != tt.want {
				t.Errorf("type(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestUnknownCommand(t *testing.T) {
	_, ok := GetCommand("nonexistent")
	if ok {
		t.Error("expected nonexistent command to not be found in registry")
	}
}
