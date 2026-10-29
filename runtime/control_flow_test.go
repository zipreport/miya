package runtime

import "testing"

// TestLoopControlErrorMethods tests LoopControlError methods
func TestLoopControlErrorMethods(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		breakErr := NewBreakError()
		errStr := breakErr.Error()
		if errStr != "break statement" {
			t.Errorf("Expected 'break statement', got %q", errStr)
		}

		continueErr := NewContinueError()
		errStr = continueErr.Error()
		if errStr != "continue statement" {
			t.Errorf("Expected 'continue statement', got %q", errStr)
		}
	})
}
