// Package executor runs shell commands and captures their result.
package executor

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"runtime"
	"time"
)

// ProcessExecutionResult holds the outcome of a shell command execution.
//
// It mirrors the original Java net.benelog.shellproxy.executor.ProcessExecutionResult.
type ProcessExecutionResult struct {
	ExitCode       int    `json:"exitCode"`
	StandardOutput string `json:"standardOutput"`
	StandardError  string `json:"standardError"`
	// TimedOut reports whether the command was killed because it exceeded the timeout.
	TimedOut bool `json:"timedOut"`
}

// ShellExecutor executes a single command line via the platform shell.
//
// A zero Timeout means "no timeout". Unlike the original Java version, which
// tokenized the command with commons-exec's CommandLine.parse, this runs the
// command through the system shell ("sh -c" / "cmd /C") so that pipes,
// redirection and other shell features work as a user would expect from a
// terminal UI.
type ShellExecutor struct {
	Timeout time.Duration
}

// New returns a ShellExecutor with the given timeout (0 disables it).
func New(timeout time.Duration) *ShellExecutor {
	return &ShellExecutor{Timeout: timeout}
}

// shellCommand returns the shell and arguments used to run a command line.
func shellCommand(command string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", command}
	}
	return "sh", []string{"-c", command}
}

// Execute runs the command and returns its captured result.
//
// A non-zero exit code is not treated as an error: it is reported in the
// returned result. Only failures to start the process yield a Go error.
func (e *ShellExecutor) Execute(command string) (ProcessExecutionResult, error) {
	ctx := context.Background()
	if e.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, e.Timeout)
		defer cancel()
	}

	name, args := shellCommand(command)
	cmd := exec.CommandContext(ctx, name, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	result := ProcessExecutionResult{
		StandardOutput: stdout.String(),
		StandardError:  stderr.String(),
		ExitCode:       cmd.ProcessState.ExitCode(),
		TimedOut:       errors.Is(ctx.Err(), context.DeadlineExceeded),
	}

	// A command that exits non-zero surfaces as *exec.ExitError; that is an
	// expected outcome, not a start failure, so do not propagate it.
	var exitErr *exec.ExitError
	if runErr != nil && !errors.As(runErr, &exitErr) {
		// The process could not be started at all (e.g. shell missing).
		return result, runErr
	}

	return result, nil
}
