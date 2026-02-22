package main

import (
	"strings"
	"testing"
)

// setupTestTrie replaces the global commandTrie with a trie containing the
// given words. Returns a cleanup function that restores the original trie.
func setupTestTrie(t *testing.T, words []string) {
	t.Helper()
	old := commandTrie
	commandTrie = newTrie()
	for _, w := range words {
		commandTrie.Insert(w)
	}
	t.Cleanup(func() { commandTrie = old })
}

func TestCompleterDo(t *testing.T) {
	tests := []struct {
		name       string
		trieWords  []string
		line       string
		pos        int
		wantSuffix string   // expected single suffix if non-empty
		wantNil    bool     // true when we expect nil result (bell or listing)
		wantStderr string   // substring expected on stderr (e.g. bell)
		wantStdout string   // substring expected on stdout (e.g. listing)
		tabCount   int      // how many times to press TAB (default 1)
	}{
		{
			name:       "no matches rings bell",
			trieWords:  []string{"echo", "exit"},
			line:       "zzz",
			pos:        3,
			wantNil:    true,
			wantStderr: "\x07",
		},
		{
			name:       "single match returns suffix with space",
			trieWords:  []string{"echo", "exit", "type"},
			line:       "ec",
			pos:        2,
			wantSuffix: "ho ",
		},
		{
			name:       "multiple matches first TAB rings bell",
			trieWords:  []string{"echo", "exit"},
			line:       "e",
			pos:        1,
			wantNil:    true,
			wantStderr: "\x07",
		},
		{
			name:       "multiple matches second TAB prints listing",
			trieWords:  []string{"echo", "exit"},
			line:       "e",
			pos:        1,
			tabCount:   2,
			wantNil:    true,
			wantStdout: "echo  exit",
		},
		{
			name:       "LCP completion returns LCP suffix",
			trieWords:  []string{"test", "testing", "tested"},
			line:       "te",
			pos:        2,
			wantSuffix: "st",
		},
		{
			name:      "prefix with space returns nil immediately",
			trieWords: []string{"echo", "exit"},
			line:      "echo ",
			pos:       5,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestTrie(t, tt.trieWords)

			comp := &builtinCompleter{}
			tabs := tt.tabCount
			if tabs == 0 {
				tabs = 1
			}

			var result [][]rune
			var length int
			for i := 0; i < tabs; i++ {
				if tt.wantStderr != "" || tt.wantStdout != "" {
					// Capture stderr/stdout on the last TAB press.
					if i == tabs-1 {
						stderrOut := captureStderr(t, func() {
							stdoutOut := captureStdout(t, func() {
								result, length = comp.Do([]rune(tt.line), tt.pos)
							})
							if tt.wantStdout != "" && !strings.Contains(stdoutOut, tt.wantStdout) {
								t.Errorf("stdout = %q, want substring %q", stdoutOut, tt.wantStdout)
							}
						})
						if tt.wantStderr != "" && !strings.Contains(stderrOut, tt.wantStderr) {
							t.Errorf("stderr = %q, want substring %q", stderrOut, tt.wantStderr)
						}
					} else {
						captureStderr(t, func() {
							captureStdout(t, func() {
								result, length = comp.Do([]rune(tt.line), tt.pos)
							})
						})
					}
				} else {
					result, length = comp.Do([]rune(tt.line), tt.pos)
				}
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil result, got %v", result)
				}
				return
			}

			if tt.wantSuffix != "" {
				if len(result) != 1 {
					t.Fatalf("expected 1 result, got %d", len(result))
				}
				got := string(result[0])
				if got != tt.wantSuffix {
					t.Errorf("suffix = %q, want %q", got, tt.wantSuffix)
				}
				if length != tt.pos {
					t.Errorf("length = %d, want %d", length, tt.pos)
				}
			}
		})
	}
}
