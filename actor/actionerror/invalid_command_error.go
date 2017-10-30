package actionerror

import "fmt"

// InvalidCommandError represents an error that happens when help is called
// with an invalid command.
type InvalidCommandError struct {
	CommandName string
}

func (err InvalidCommandError) Error() string {
	return fmt.Sprintf("'%s' is not a registered command. See 'cf help -a'", err.CommandName)
}
