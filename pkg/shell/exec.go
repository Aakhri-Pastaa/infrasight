// Package shell provides safe, read-only command execution for probes.
//
// InfraSight never mutates the host. Every command run through this package is
// expected to be non-destructive (status/list/show style invocations). The
// timeout guards against probes that hang on a wedged subsystem.
package shell

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// Run executes name with args under a timeout and returns trimmed stdout.
// On failure it returns whatever stdout was captured plus a *CommandError that
// includes stderr, so callers can degrade gracefully on partial output.
func Run(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := strings.TrimRight(stdout.String(), "\n")
	if err != nil {
		return out, &CommandError{Name: name, Err: err, Stderr: strings.TrimSpace(stderr.String())}
	}
	return out, nil
}

// Available reports whether a binary is resolvable in PATH.
func Available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// CommandError wraps a failed command invocation with its stderr.
type CommandError struct {
	Name   string
	Err    error
	Stderr string
}

func (e *CommandError) Error() string {
	if e.Stderr != "" {
		return e.Name + ": " + e.Err.Error() + ": " + e.Stderr
	}
	return e.Name + ": " + e.Err.Error()
}

func (e *CommandError) Unwrap() error { return e.Err }
