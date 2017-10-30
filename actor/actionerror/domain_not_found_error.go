package actionerror

import "fmt"

// DomainNotFoundError is an error wrapper that represents the case
// when the domain is not found.
type DomainNotFoundError struct {
	Name string
	GUID string
}

// Error method to display the error message.
func (e DomainNotFoundError) Error() string {
	switch {
	case e.Name != "":
		return fmt.Sprintf("Domain %s not found", e.Name)
	case e.GUID != "":
		return fmt.Sprintf("Domain with GUID %s not found", e.GUID)
	default:
		return "Domain not found"
	}
}
