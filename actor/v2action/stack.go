package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type Stack ccv2.Stack

// StackNotFoundError is returned when a requested stack is not found.
type StackNotFoundError struct {
	GUID string
}

func (e StackNotFoundError) Error() string {
	return fmt.Sprintf("Stack with GUID '%s' not found.", e.GUID)
}

// GetStack returns the stack information associated with the provided stack GUID.
func (actor Actor) GetStack(guid string) (Stack, Warnings, error) {
	stack, warnings, err := actor.CloudControllerClient.GetStack(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return Stack{}, Warnings(warnings), StackNotFoundError{GUID: guid}
	}

	return Stack(stack), Warnings(warnings), err
}
