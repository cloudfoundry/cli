package ui

import (
	"fmt"
)

type SkipStageError struct {
	cause       error
	skipMessage string
}

func NewSkipStageError(cause error, skipMessage string) SkipStageError {
	return SkipStageError{
		cause:       cause,
		skipMessage: skipMessage,
	}
}

func (e SkipStageError) Error() string {
	return fmt.Sprintf("%s: %s", e.skipMessage, e.cause)
}

func (e SkipStageError) SkipMessage() string {
	return e.skipMessage
}

func (e SkipStageError) Cause() error {
	return e.cause
}
