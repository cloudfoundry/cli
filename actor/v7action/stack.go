package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

func (actor *Actor) GetStackByName(stackName string) (resources.Stack, Warnings, error) {
	stacks, warnings, err := actor.CloudControllerClient.GetStacks(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{stackName}},
	)

	if err != nil {
		return resources.Stack{}, Warnings(warnings), err
	}

	if len(stacks) == 0 {
		return resources.Stack{}, Warnings(warnings), actionerror.StackNotFoundError{Name: stackName}
	}

	return resources.Stack(stacks[0]), Warnings(warnings), nil
}

func (actor Actor) GetStacks(labelSelector string) ([]resources.Stack, Warnings, error) {
	var (
		stacks   []resources.Stack
		warnings ccv3.Warnings
		err      error
	)
	if len(labelSelector) > 0 {
		queries := []ccv3.Query{
			ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{labelSelector}},
		}

		stacks, warnings, err = actor.CloudControllerClient.GetStacks(queries...)
	} else {
		stacks, warnings, err = actor.CloudControllerClient.GetStacks()
	}

	if err != nil {
		return nil, Warnings(warnings), err
	}

	return stacks, Warnings(warnings), nil
}
