package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// --- Types ---

type proc struct {
	cmd  *exec.Cmd
	done chan struct{}
}

type pipeline struct {
	n     int        // number of segments
	pipeR []*os.File // read ends between segments
	pipeW []*os.File // write ends between segments
	procs []proc
}

// --- Entry point ---

func executePipeline(segments []string) {
	p := &pipeline{n: len(segments)}
	if err := p.createPipes(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var cleanups []func()
	defer func() {
		for _, cl := range cleanups {
			cl()
		}
	}()

	for i, seg := range segments {
		cl, err := p.startSegment(i, seg)
		if cl != nil {
			cleanups = append(cleanups, cl)
		}
		if err != nil {
			p.closePipes()
			return
		}
	}

	p.wait()
}

// --- Core segment execution ---

// startSegment parses, wires I/O, and launches segment i. It returns an
// optional redirect cleanup function and any error that prevents execution.
func (p *pipeline) startSegment(i int, seg string) (cleanup func(), err error) {
	parsed, err := parseCommand(seg)
	if err != nil {
		return nil, err
	}

	stdin, stdout, stderr := p.segmentIO(i)

	// Apply redirections (typically only on the last segment).
	if len(parsed.Redirects) > 0 {
		rOut, rErr, cl, err := openRedirects(parsed.Redirects)
		if err != nil {
			return nil, err
		}
		cleanup = cl
		if rOut != os.Stdout {
			stdout = rOut
		}
		if rErr != os.Stderr {
			stderr = rErr
		}
	}

	if builtin, ok := GetCommand(parsed.Name); ok {
		p.startBuiltin(i, builtin, parsed.Args, stdout, stderr)
		return cleanup, nil
	}

	if err := p.startExternal(i, parsed.Name, parsed.Args, stdin, stdout, stderr); err != nil {
		return cleanup, err
	}
	return cleanup, nil
}

func (p *pipeline) startBuiltin(i int, builtin Command, args []string, stdout, stderr *os.File) {
	done := make(chan struct{})
	p.procs[i] = proc{done: done}
	go func() {
		defer close(done)
		origOut, origErr := os.Stdout, os.Stderr
		os.Stdout = stdout
		os.Stderr = stderr
		builtin.Run(args)
		os.Stdout = origOut
		os.Stderr = origErr
		p.closeParentEnds(i)
	}()
}

func (p *pipeline) startExternal(i int, name string, args []string, stdin, stdout, stderr *os.File) error {
	c := exec.Command(name, args...)
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr
	if err := c.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			fmt.Printf("%s: command not found\n", name)
		}
		return err
	}
	p.procs[i] = proc{cmd: c}
	p.closeParentEnds(i)
	return nil
}

// --- Pipe plumbing ---

func (p *pipeline) createPipes() error {
	p.pipeR = make([]*os.File, p.n-1)
	p.pipeW = make([]*os.File, p.n-1)
	p.procs = make([]proc, p.n)
	for i := 0; i < p.n-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			p.closePipes()
			return err
		}
		p.pipeR[i] = r
		p.pipeW[i] = w
	}
	return nil
}

func (p *pipeline) closePipes() {
	for i := range p.pipeR {
		if p.pipeR[i] != nil {
			p.pipeR[i].Close()
		}
		if p.pipeW[i] != nil {
			p.pipeW[i].Close()
		}
	}
}

// segmentIO returns the stdin/stdout/stderr for segment i based on pipe wiring.
func (p *pipeline) segmentIO(i int) (stdin, stdout, stderr *os.File) {
	stdin = os.Stdin
	stdout = os.Stdout
	stderr = os.Stderr
	if i > 0 {
		stdin = p.pipeR[i-1]
	}
	if i < p.n-1 {
		stdout = p.pipeW[i]
	}
	return
}

// closeParentEnds closes the parent process's copies of pipe ends that
// segment i now owns (via fork or goroutine).
func (p *pipeline) closeParentEnds(i int) {
	if i < p.n-1 {
		p.pipeW[i].Close()
		p.pipeW[i] = nil
	}
	if i > 0 {
		p.pipeR[i-1].Close()
		p.pipeR[i-1] = nil
	}
}

func (p *pipeline) wait() {
	for _, pr := range p.procs {
		if pr.cmd != nil {
			pr.cmd.Wait()
		} else if pr.done != nil {
			<-pr.done
		}
	}
}
