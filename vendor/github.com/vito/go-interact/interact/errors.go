package interact

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrNotANumber is used internally by Resolve when the user enters a bogus
// value when resolving into an int.
//
// Resolve will retry on this error; it is only exposed so you can know where
// the string is coming from.
var ErrNotANumber = errors.New("not a number")

// ErrNotBoolean is used internally by Resolve when the user enters a bogus
// value when resolving into a bool.
//
// Resolve will retry on this error; it is only exposed so you can know where
// the string is coming from.
var ErrNotBoolean = errors.New("not y, n, yes, or no")

// NotAssignableError is returned by Resolve when the value present in the
// Choice the user selected is not assignable to the destination value during
// Resolve.
type NotAssignableError struct {
	Destination reflect.Type
	Value       reflect.Type
}

func (err NotAssignableError) Error() string {
	return fmt.Sprintf("chosen value (%T) is not assignable to %T", err.Value, err.Destination)
}
