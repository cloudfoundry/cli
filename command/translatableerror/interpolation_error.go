package translatableerror

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

func (e InterpolationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
