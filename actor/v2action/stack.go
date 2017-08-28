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
	Name string
}

func (e StackNotFoundError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("Stack with GUID '%s' not found.", e.GUID)
	}

	return fmt.Sprintf("Stack '%s' not found.", e.Name)
}

// GetStack returns the stack information associated with the provided stack GUID.
func (actor Actor) GetStack(guid string) (Stack, Warnings, error) {
	stack, warnings, err := actor.CloudControllerClient.GetStack(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return Stack{}, Warnings(warnings), StackNotFoundError{GUID: guid}
	}

	return Stack(stack), Warnings(warnings), err
}

// GetStackByName returns the provided stack
func (actor Actor) GetStackByName(stackName string) (Stack, Warnings, error) {
	stacks, warnings, err := actor.CloudControllerClient.GetStacks(ccv2.Query{
		Filter:   ccv2.NameFilter,
		Operator: ccv2.EqualOperator,
		Values:   []string{stackName},
	})
	if err != nil {
		return Stack{}, Warnings(warnings), err
	}

	if len(stacks) == 0 {
		return Stack{}, Warnings(warnings), StackNotFoundError{Name: stackName}
	}

	return Stack(stacks[0]), Warnings(warnings), nil
}
