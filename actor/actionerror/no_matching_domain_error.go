package actionerror

import "fmt"

type NoMatchingDomainError struct {
	Route string
}

func (e NoMatchingDomainError) Error() string {
	return fmt.Sprintln("Unable to find a matching domain for", e.Route)
}
