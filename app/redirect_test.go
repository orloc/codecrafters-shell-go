package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenRedirects(t *testing.T) {
	t.Run("no redirects returns original stdout and stderr", func(t *testing.T) {
		origStdout := os.Stdout
		origStderr := os.Stderr
		stdout, stderr, cleanup, err := openRedirects(nil)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
		if stdout != origStdout {
			t.Error("expected stdout to be os.Stdout")
		}
		if stderr != origStderr {
			t.Error("expected stderr to be os.Stderr")
		}
	})

	t.Run("stdout truncate redirect creates file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "out.txt")
		redirects := []Redirect{{Fd: 1, Op: ">", File: path}}

		stdout, stderr, cleanup, err := openRedirects(redirects)
		if err != nil {
			t.Fatal(err)
		}
		origStderr := os.Stderr
		if stderr != origStderr {
			t.Error("expected stderr to remain os.Stderr")
		}
		stdout.WriteString("hello\n")
		cleanup()

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "hello\n" {
			t.Errorf("file content = %q, want %q", string(data), "hello\n")
		}
	})

	t.Run("stderr redirect creates file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "err.txt")
		redirects := []Redirect{{Fd: 2, Op: ">", File: path}}

		_, stderr, cleanup, err := openRedirects(redirects)
		if err != nil {
			t.Fatal(err)
		}
		stderr.WriteString("oops\n")
		cleanup()

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "oops\n" {
			t.Errorf("file content = %q, want %q", string(data), "oops\n")
		}
	})

	t.Run("append mode appends to existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "out.txt")
		os.WriteFile(path, []byte("first\n"), 0644)

		redirects := []Redirect{{Fd: 1, Op: ">>", File: path}}
		stdout, _, cleanup, err := openRedirects(redirects)
		if err != nil {
			t.Fatal(err)
		}
		stdout.WriteString("second\n")
		cleanup()

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "first\nsecond\n" {
			t.Errorf("file content = %q, want %q", string(data), "first\nsecond\n")
		}
	})

	t.Run("invalid file path returns error", func(t *testing.T) {
		redirects := []Redirect{{Fd: 1, Op: ">", File: "/no/such/dir/file.txt"}}
		_, _, _, err := openRedirects(redirects)
		if err == nil {
			t.Error("expected error for invalid file path")
		}
	})

	t.Run("cleanup restores original stdout and stderr", func(t *testing.T) {
		origStdout := os.Stdout
		origStderr := os.Stderr
		dir := t.TempDir()

		redirects := []Redirect{
			{Fd: 1, Op: ">", File: filepath.Join(dir, "out.txt")},
			{Fd: 2, Op: ">", File: filepath.Join(dir, "err.txt")},
		}
		_, _, cleanup, err := openRedirects(redirects)
		if err != nil {
			t.Fatal(err)
		}
		cleanup()

		if os.Stdout != origStdout {
			t.Error("expected os.Stdout to be restored after cleanup")
		}
		if os.Stderr != origStderr {
			t.Error("expected os.Stderr to be restored after cleanup")
		}
	})
}
