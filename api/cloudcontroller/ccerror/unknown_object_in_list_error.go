package ccerror

import (
	"fmt"
	"reflect"
)

// UnknownObjectInListError is returned when iterating through a paginated
// list. Assuming tests are written for the paginated function, this should be
// impossible to get.
type UnknownObjectInListError struct {
	Expected   interface{}
	Unexpected interface{}
}

func (e UnknownObjectInListError) Error() string {
	return fmt.Sprintf(
		"Error while processing a paginated list. Expected %s but %s was returned",
		reflect.TypeOf(e.Expected),
		reflect.TypeOf(e.Unexpected),
	)
}
