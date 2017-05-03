package ccerror

import "fmt"

// NilObjectError gets returned when passed a nil object as a parameter.
type NilObjectError struct {
	Object string
}

func (e NilObjectError) Error() string {
	return fmt.Sprintf("%s cannot be nil", e.Object)
}
