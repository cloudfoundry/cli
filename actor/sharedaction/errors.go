package sharedaction

import "fmt"

type ErrorInvalidCommand struct {
	CommandName string
}

func (err ErrorInvalidCommand) Error() string {
	return fmt.Sprintf("'%s' is not a registered command. See 'cf help'", err.CommandName)
}
