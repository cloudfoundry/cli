package ccerror

import "fmt"

// UserNotFoundError is returned when a role does not exist.
type UserNotFoundError struct {
	Username string
	Origin   string
	IsClient   bool
}

func (e UserNotFoundError) Error() string {
	// return "Invalid user. Ensure User exists and you have access to it."
	var baseString, originString, nameString string
	if e.IsClient {
		baseString = "Client not found"
	} else {
		baseString = "User not found"
	}
	originString = "."
	if e.Origin != "" {
		originString = fmt.Sprintf(" and origin %s", e.Origin)
	}
	if e.Username != "" {
		nameString = fmt.Sprintf(" with name %s", e.Username)
	}
	return fmt.Sprintf("%s%s%s", baseString, nameString, originString)
}
