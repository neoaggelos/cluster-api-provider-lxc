package incus

import "errors"

type terminalError struct {
	error
}

// IsTerminalError checks whether the error is a terminalError.
// These are returned to indicate non-retriable errors.
func IsTerminalError(err error) bool {
	return errors.As(err, &terminalError{})
}
