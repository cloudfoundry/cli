package ccerror

import "fmt"

// UserNotFoundError is returned when a role does not exist.
type UserNotFoundError struct {
	Username string
	Origin   string
}

func (e UserNotFoundError) Error() string {
	originString := "."
	if e.Origin != "" {
		originString = fmt.Sprintf(" and origin %s", e.Origin)
	}
	return fmt.Sprintf("No user exists with the username %s%s", e.Username, originString)
}
