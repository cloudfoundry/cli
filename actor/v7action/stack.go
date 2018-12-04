package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Stack ccv3.Stack

func (actor *Actor) GetStackByName(stackName string) (Stack, Warnings, error) {
	stacks, warnings, err := actor.CloudControllerClient.GetStacks(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{stackName}},
	)

	if err != nil {
		return Stack{}, Warnings(warnings), err
	}

	if len(stacks) == 0 {
		return Stack{}, Warnings(warnings), actionerror.StackNotFoundError{Name: stackName}
	}

	return Stack(stacks[0]), Warnings(warnings), nil
}

func (actor Actor) GetStacks() ([]Stack, Warnings, error) {
	ccv3Stacks, warnings, err := actor.CloudControllerClient.GetStacks()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	var stacks []Stack
	for _, stack := range ccv3Stacks {
		stacks = append(stacks, Stack(stack))
	}
	return stacks, Warnings(warnings), nil
}
