package actionerror

import "fmt"

// EnvironmentVariableNotSetError is returned when trying to unset env variable
// that was not previously set.
type EnvironmentVariableNotSetError struct {
	EnvironmentVariableName string
}

func (e EnvironmentVariableNotSetError) Error() string {
	return fmt.Sprintf("Env variable %s was not set.", e.EnvironmentVariableName)
}
