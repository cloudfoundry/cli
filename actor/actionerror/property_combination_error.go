package actionerror

import (
	"fmt"
	"strings"
)

type PropertyCombinationError struct {
	AppName    string
	Properties []string
}

func (e PropertyCombinationError) Error() string {
	return fmt.Sprintln("Cannot use the following properties together:", strings.Join(e.Properties, ", "))
}
