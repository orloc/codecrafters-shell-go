package main

import (
	"os"
	"strings"
	"testing"
)

func TestSegmentIO(t *testing.T) {
	p := &pipeline{n: 3}
	if err := p.createPipes(); err != nil {
		t.Fatal(err)
	}
	defer p.closePipes()

	t.Run("first segment reads from stdin writes to pipe", func(t *testing.T) {
		stdin, stdout, stderr := p.segmentIO(0)
		if stdin != os.Stdin {
			t.Error("expected stdin to be os.Stdin")
		}
		if stdout != p.pipeW[0] {
			t.Error("expected stdout to be pipeW[0]")
		}
		if stderr != os.Stderr {
			t.Error("expected stderr to be os.Stderr")
		}
	})

	t.Run("middle segment reads from pipe writes to pipe", func(t *testing.T) {
		stdin, stdout, stderr := p.segmentIO(1)
		if stdin != p.pipeR[0] {
			t.Error("expected stdin to be pipeR[0]")
		}
		if stdout != p.pipeW[1] {
			t.Error("expected stdout to be pipeW[1]")
		}
		if stderr != os.Stderr {
			t.Error("expected stderr to be os.Stderr")
		}
	})

	t.Run("last segment reads from pipe writes to stdout", func(t *testing.T) {
		stdin, stdout, stderr := p.segmentIO(2)
		if stdin != p.pipeR[1] {
			t.Error("expected stdin to be pipeR[1]")
		}
		if stdout != os.Stdout {
			t.Error("expected stdout to be os.Stdout")
		}
		if stderr != os.Stderr {
			t.Error("expected stderr to be os.Stderr")
		}
	})
}

func TestExecutePipeline(t *testing.T) {
	tests := []struct {
		name     string
		segments []string
		want     string
	}{
		{
			name:     "builtin to external",
			segments: []string{"echo hello", " cat"},
			want:     "hello",
		},
		{
			name:     "three stage pipeline",
			segments: []string{"echo hello", " cat", " cat"},
			want:     "hello",
		},
		{
			name:     "pipe to wc",
			segments: []string{"echo hello world", " wc -w"},
			want:     "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := captureStdout(t, func() {
				executePipeline(tt.segments)
			})
			got = strings.TrimSpace(got)
			want := strings.TrimSpace(tt.want)
			if got != want {
				t.Errorf("executePipeline() output = %q, want %q", got, want)
			}
		})
	}
}
