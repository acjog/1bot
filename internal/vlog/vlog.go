// Package vlog formats verbose protocol markers for a terminal: a blank line
// separates each marker from the child process's own stderr output, and the
// [tag] marker is colored red so it stands out. Every -v frame/marker log goes
// through Frame so the two directions stay visually consistent.
package vlog

import (
	"fmt"
	"io"
)

const (
	red   = "\x1b[31m"
	reset = "\x1b[0m"
)

// Frame prints one verbose marker line: two blank lines, a red [tag], then the
// payload on its own line.
func Frame(w io.Writer, tag string, payload string) {
	fmt.Fprintf(w, "\n\n%s[%s]%s %s\n", red, tag, reset, payload)
}
