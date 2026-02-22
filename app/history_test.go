package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecord(t *testing.T) {
	h := NewHistory()

	h.Record("echo hello")
	h.Record("ls -la")

	if len(h.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(h.entries))
	}
	if h.entries[0] != "echo hello" {
		t.Errorf("entry 0 = %q, want %q", h.entries[0], "echo hello")
	}
	if h.entries[1] != "ls -la" {
		t.Errorf("entry 1 = %q, want %q", h.entries[1], "ls -la")
	}
}

func TestReadFile(t *testing.T) {
	t.Run("reads non-empty lines", func(t *testing.T) {
		h := NewHistory()
		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("echo hello\necho world\n\n"), 0644)

		if err := h.ReadFile(path); err != nil {
			t.Fatal(err)
		}
		if len(h.entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(h.entries))
		}
		if h.entries[0] != "echo hello" {
			t.Errorf("entry 0 = %q, want %q", h.entries[0], "echo hello")
		}
		if h.entries[1] != "echo world" {
			t.Errorf("entry 1 = %q, want %q", h.entries[1], "echo world")
		}
	})

	t.Run("appends to existing history", func(t *testing.T) {
		h := NewHistory()
		h.Record("first")
		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("second\n"), 0644)

		if err := h.ReadFile(path); err != nil {
			t.Fatal(err)
		}
		if len(h.entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(h.entries))
		}
		if h.entries[0] != "first" {
			t.Errorf("entry 0 = %q, want %q", h.entries[0], "first")
		}
		if h.entries[1] != "second" {
			t.Errorf("entry 1 = %q, want %q", h.entries[1], "second")
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		h := NewHistory()
		if err := h.ReadFile("/nonexistent/path"); err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func TestWriteFile(t *testing.T) {
	t.Run("writes all entries with trailing newline", func(t *testing.T) {
		h := NewHistory()
		h.Record("echo hello")
		h.Record("echo world")

		path := filepath.Join(t.TempDir(), "history")
		if err := h.WriteFile(path); err != nil {
			t.Fatal(err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		want := "echo hello\necho world\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})

	t.Run("creates file if it does not exist", func(t *testing.T) {
		h := NewHistory()
		h.Record("cmd")
		path := filepath.Join(t.TempDir(), "new_history")
		if err := h.WriteFile(path); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file to be created: %v", err)
		}
	})

	t.Run("empty history writes empty file", func(t *testing.T) {
		h := NewHistory()
		path := filepath.Join(t.TempDir(), "history")
		if err := h.WriteFile(path); err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(path)
		if len(data) != 0 {
			t.Errorf("expected empty file, got %q", string(data))
		}
	})
}

func TestAppendFile(t *testing.T) {
	t.Run("appends to existing file", func(t *testing.T) {
		h := NewHistory()
		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("old command\n"), 0644)

		h.Record("new command")
		if err := h.AppendFile(path); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		want := "old command\nnew command\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})

	t.Run("creates file if it does not exist", func(t *testing.T) {
		h := NewHistory()
		h.Record("cmd")
		path := filepath.Join(t.TempDir(), "new_history")
		if err := h.AppendFile(path); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		want := "cmd\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})

	t.Run("second append only writes new entries", func(t *testing.T) {
		h := NewHistory()
		path := filepath.Join(t.TempDir(), "history")

		h.Record("first")
		h.AppendFile(path)

		h.Record("second")
		h.AppendFile(path)

		data, _ := os.ReadFile(path)
		want := "first\nsecond\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})

	t.Run("empty history appends nothing", func(t *testing.T) {
		h := NewHistory()
		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("existing\n"), 0644)

		if err := h.AppendFile(path); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		want := "existing\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})
}

func TestPrint(t *testing.T) {
	t.Run("prints all entries", func(t *testing.T) {
		h := NewHistory()
		h.Record("echo hello")
		h.Record("echo world")

		got := captureStdout(t, func() { h.Print(0) })
		want := "    1  echo hello\n    2  echo world\n"
		if got != want {
			t.Errorf("Print(0) = %q, want %q", got, want)
		}
	})

	t.Run("prints last n entries", func(t *testing.T) {
		h := NewHistory()
		h.Record("first")
		h.Record("second")
		h.Record("third")

		got := captureStdout(t, func() { h.Print(2) })
		want := "    2  second\n    3  third\n"
		if got != want {
			t.Errorf("Print(2) = %q, want %q", got, want)
		}
	})

	t.Run("n larger than history prints all", func(t *testing.T) {
		h := NewHistory()
		h.Record("only")

		got := captureStdout(t, func() { h.Print(10) })
		want := "    1  only\n"
		if got != want {
			t.Errorf("Print(10) = %q, want %q", got, want)
		}
	})
}
