// redirect.go — file-based I/O redirection (>, >>, 1>, 2>).
//
// openRedirects opens target files and returns replacement stdout/stderr
// writers plus a cleanup function that closes files and restores originals.
package main

import "os"

// Redirect describes a single I/O redirection (e.g. "> file", "2>> err.log").
type Redirect struct {
	Fd   int    // 1 = stdout, 2 = stderr
	Op   string // ">" or ">>"
	File string // target file path
}

// openRedirects opens files for each redirect and returns the stdout/stderr
// writers to use. The returned cleanup function closes all opened files and
// restores os.Stdout/os.Stderr to their original values.
func openRedirects(redirects []Redirect) (stdout, stderr *os.File, cleanup func(), err error) {
	origStdout := os.Stdout
	origStderr := os.Stderr
	stdout = origStdout
	stderr = origStderr

	var files []*os.File
	for _, r := range redirects {
		var f *os.File
		if r.Op == ">" {
			f, err = os.Create(r.File)
		} else {
			f, err = os.OpenFile(r.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		}
		if err != nil {
			for _, cf := range files {
				cf.Close()
			}
			return nil, nil, nil, err
		}
		files = append(files, f)
		if r.Fd == 1 {
			stdout = f
		} else {
			stderr = f
		}
	}

	cleanup = func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		for _, f := range files {
			f.Close()
		}
	}

	return stdout, stderr, cleanup, nil
}
