package magic

import (
	"fmt"
)

// Error represents an error originating from the underlying Magic library.
type Error struct {
	Errno   int    // The value of errno, if any.
	Message string // The actual error message.
}

// Error returns a descriptive error message.
func (e *Error) Error() string {
	return fmt.Sprintf("magic: %s", e.Message)
}
