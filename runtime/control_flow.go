package runtime

import "fmt"

// LoopControlError represents break and continue flow control
type LoopControlError struct {
	Type string // "break" or "continue"
}

func (e *LoopControlError) Error() string {
	return fmt.Sprintf("%s statement", e.Type)
}

// IsBreak returns true if this is a break control error
func (e *LoopControlError) IsBreak() bool {
	return e.Type == "break"
}

// IsContinue returns true if this is a continue control error
func (e *LoopControlError) IsContinue() bool {
	return e.Type == "continue"
}

// NewBreakError creates a break control error
func NewBreakError() *LoopControlError {
	return &LoopControlError{Type: "break"}
}

// NewContinueError creates a continue control error
func NewContinueError() *LoopControlError {
	return &LoopControlError{Type: "continue"}
}
