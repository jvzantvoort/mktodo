package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

// Exit codes, as defined in the specification (NFR-2).
const (
	ExitOK      = 0
	ExitGeneral = 1
	ExitUsage   = 2 // invalid usage, unknown project or item
	ExitConfig  = 3 // missing or invalid configuration
	ExitFile    = 4 // the TODO file can't be read or written
	ExitGit     = 5 // not in a git repository
)

// exitError is an error with the process exit code it should produce.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

func withCode(code int, err error) error {
	if err == nil {
		return nil
	}
	return &exitError{code: code, err: err}
}

func usageError(err error) error  { return withCode(ExitUsage, err) }
func configError(err error) error { return withCode(ExitConfig, err) }
func fileError(err error) error   { return withCode(ExitFile, err) }
func gitError(err error) error    { return withCode(ExitGit, err) }

// exitCode returns the exit code for an error returned by a command.
func exitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var e *exitError
	if errors.As(err, &e) {
		return e.code
	}
	return ExitGeneral
}

// usageArgs wraps a cobra argument validator so its errors are usage errors.
func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		return usageError(validate(cmd, args))
	}
}

// formatError returns the text shown for an error. Messages from the
// messages package already start with "Error:"; cobra's own errors don't.
func formatError(err error) string {
	msg := err.Error()
	if strings.HasPrefix(msg, "Error:") || strings.HasPrefix(msg, "Operation cancelled") {
		return msg
	}
	return "Error: " + msg
}
