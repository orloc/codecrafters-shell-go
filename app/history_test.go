package main

import (
	"os"
	"path/filepath"
	"testing"
)

// resetHistory clears the global commandHistory for test isolation.
func resetHistory() {
	commandHistory = nil
}

func TestRecordHistory(t *testing.T) {
	resetHistory()
	defer resetHistory()

	recordHistory("echo hello")
	recordHistory("ls -la")

	if len(commandHistory) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(commandHistory))
	}
	if commandHistory[0] != "echo hello" {
		t.Errorf("entry 0 = %q, want %q", commandHistory[0], "echo hello")
	}
	if commandHistory[1] != "ls -la" {
		t.Errorf("entry 1 = %q, want %q", commandHistory[1], "ls -la")
	}
}

func TestReadHistoryFile(t *testing.T) {
	t.Run("reads non-empty lines", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("echo hello\necho world\n\n"), 0644)

		if err := readHistoryFile(path); err != nil {
			t.Fatal(err)
		}
		if len(commandHistory) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(commandHistory))
		}
		if commandHistory[0] != "echo hello" {
			t.Errorf("entry 0 = %q, want %q", commandHistory[0], "echo hello")
		}
		if commandHistory[1] != "echo world" {
			t.Errorf("entry 1 = %q, want %q", commandHistory[1], "echo world")
		}
	})

	t.Run("appends to existing history", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		recordHistory("first")
		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("second\n"), 0644)

		if err := readHistoryFile(path); err != nil {
			t.Fatal(err)
		}
		if len(commandHistory) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(commandHistory))
		}
		if commandHistory[0] != "first" {
			t.Errorf("entry 0 = %q, want %q", commandHistory[0], "first")
		}
		if commandHistory[1] != "second" {
			t.Errorf("entry 1 = %q, want %q", commandHistory[1], "second")
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		if err := readHistoryFile("/nonexistent/path"); err == nil {
			t.Error("expected error for missing file")
		}
	})
}

func TestWriteHistoryFile(t *testing.T) {
	t.Run("writes all entries with trailing newline", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		recordHistory("echo hello")
		recordHistory("echo world")

		path := filepath.Join(t.TempDir(), "history")
		if err := writeHistoryFile(path); err != nil {
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
		resetHistory()
		defer resetHistory()

		recordHistory("cmd")
		path := filepath.Join(t.TempDir(), "new_history")
		if err := writeHistoryFile(path); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file to be created: %v", err)
		}
	})

	t.Run("empty history writes empty file", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		path := filepath.Join(t.TempDir(), "history")
		if err := writeHistoryFile(path); err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(path)
		if len(data) != 0 {
			t.Errorf("expected empty file, got %q", string(data))
		}
	})
}

func TestAppendHistoryFile(t *testing.T) {
	t.Run("appends to existing file", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("old command\n"), 0644)

		recordHistory("new command")
		if err := appendHistoryFile(path); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		want := "old command\nnew command\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})

	t.Run("creates file if it does not exist", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		recordHistory("cmd")
		path := filepath.Join(t.TempDir(), "new_history")
		if err := appendHistoryFile(path); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		want := "cmd\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})

	t.Run("empty history appends nothing", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		path := filepath.Join(t.TempDir(), "history")
		os.WriteFile(path, []byte("existing\n"), 0644)

		if err := appendHistoryFile(path); err != nil {
			t.Fatal(err)
		}

		data, _ := os.ReadFile(path)
		want := "existing\n"
		if string(data) != want {
			t.Errorf("file contents = %q, want %q", string(data), want)
		}
	})
}

func TestPrintHistory(t *testing.T) {
	t.Run("prints all entries", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		recordHistory("echo hello")
		recordHistory("echo world")

		got := captureStdout(t, func() { printHistory(0) })
		want := "    1  echo hello\n    2  echo world\n"
		if got != want {
			t.Errorf("printHistory(0) = %q, want %q", got, want)
		}
	})

	t.Run("prints last n entries", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		recordHistory("first")
		recordHistory("second")
		recordHistory("third")

		got := captureStdout(t, func() { printHistory(2) })
		want := "    2  second\n    3  third\n"
		if got != want {
			t.Errorf("printHistory(2) = %q, want %q", got, want)
		}
	})

	t.Run("n larger than history prints all", func(t *testing.T) {
		resetHistory()
		defer resetHistory()

		recordHistory("only")

		got := captureStdout(t, func() { printHistory(10) })
		want := "    1  only\n"
		if got != want {
			t.Errorf("printHistory(10) = %q, want %q", got, want)
		}
	})
}
