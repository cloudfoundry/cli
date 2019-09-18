package actionerror

import (
	"fmt"
)

// UserNotFoundError is an error wrapper that represents the case
// when the user is not found in UAA.
type UAAUserNotFoundError struct {
	Username string
}

// Error method to display the error message.
func (e UAAUserNotFoundError) Error() string {
	return fmt.Sprintf("User '%s' does not exist.", e.Username)
}
