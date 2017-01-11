package sharedaction

import "fmt"

// ErrorInvalidCommand represents an error that happens when help is called
// with an invalid command.
type ErrorInvalidCommand struct {
	CommandName string
}

func (err ErrorInvalidCommand) Error() string {
	return fmt.Sprintf("'%s' is not a registered command. See 'cf help'", err.CommandName)
}
