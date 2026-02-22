package main

import (
	"os"
	"testing"
)

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

func TestCdCommand(t *testing.T) {
	cmd, ok := GetCommand("cd")
	if !ok {
		t.Fatal("cd command not found in registry")
	}

	// Save and restore working directory for each subtest.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	t.Run("cd to valid directory", func(t *testing.T) {
		dir := t.TempDir()
		os.Chdir(origDir)
		cmd.Run([]string{dir})
		wd, _ := os.Getwd()
		if wd != dir {
			t.Errorf("after cd %q, cwd = %q, want %q", dir, wd, dir)
		}
	})

	t.Run("cd to invalid directory prints error", func(t *testing.T) {
		os.Chdir(origDir)
		got := captureStdout(t, func() {
			cmd.Run([]string{"/no/such/dir"})
		})
		want := "cd: /no/such/dir: No such file or directory\n"
		if got != want {
			t.Errorf("cd error output = %q, want %q", got, want)
		}
	})

	t.Run("cd with no args goes home", func(t *testing.T) {
		os.Chdir(origDir)
		home, _ := os.UserHomeDir()
		cmd.Run(nil)
		wd, _ := os.Getwd()
		if wd != home {
			t.Errorf("after cd (no args), cwd = %q, want %q", wd, home)
		}
	})

	t.Run("cd to tilde goes home", func(t *testing.T) {
		os.Chdir(origDir)
		home, _ := os.UserHomeDir()
		cmd.Run([]string{"~"})
		wd, _ := os.Getwd()
		if wd != home {
			t.Errorf("after cd ~, cwd = %q, want %q", wd, home)
		}
	})
}

func TestPwdCommand(t *testing.T) {
	cmd, ok := GetCommand("pwd")
	if !ok {
		t.Fatal("pwd command not found in registry")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	t.Run("prints current directory", func(t *testing.T) {
		dir := t.TempDir()
		os.Chdir(dir)
		got := captureStdout(t, func() {
			cmd.Run(nil)
		})
		if got != dir+"\n" {
			t.Errorf("pwd output = %q, want %q", got, dir+"\n")
		}
	})
}
