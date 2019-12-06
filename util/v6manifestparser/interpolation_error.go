package v6manifestparser

import (
	"fmt"
	"strings"
)

type InterpolationError struct {
	Err error
}

func (e InterpolationError) Error() string {
	return fmt.Sprint(strings.Replace(e.Err.Error(), "\n", ", ", -1))
}
