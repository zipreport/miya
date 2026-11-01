package runtime

import "errors"

// ErrOutputTooLarge is wrapped by every error that rejects a result for its size.
var ErrOutputTooLarge = errors.New("result too large")

// SizeError is a size-limit error with its own message that wraps ErrOutputTooLarge.
type SizeError struct{ Msg string }

func (e *SizeError) Error() string { return e.Msg }

func (e *SizeError) Unwrap() error { return ErrOutputTooLarge }
