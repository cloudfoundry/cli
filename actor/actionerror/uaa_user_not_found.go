package actionerror

import (
	"fmt"
)

// UserNotFoundError is an error wrapper that represents the case
// when the user is not found in UAA.
type UAAUserNotFoundError struct {
	Username string
	Origin   string
}

// Error method to display the error message.
func (e UAAUserNotFoundError) Error() string {
	if e.Origin != "" && e.Origin != "uaa" {
		return fmt.Sprintf("User '%s' with origin '%s' does not exist.", e.Username, e.Origin)
	}

	return fmt.Sprintf("User '%s' does not exist.", e.Username)
}
