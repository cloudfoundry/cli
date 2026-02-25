package v7action

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
)

func (actor *Actor) GetStackByName(stackName string) (resources.Stack, Warnings, error) {
	stacks, warnings, err := actor.CloudControllerClient.GetStacks(
		ccv3.Query{Key: ccv3.NameFilter, Values: []string{stackName}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
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

func (actor Actor) UpdateStack(stackGUID string, state string, reason string) (resources.Stack, Warnings, error) {
	stack, warnings, err := actor.CloudControllerClient.UpdateStack(stackGUID, state, reason)
	if err != nil {
		return resources.Stack{}, Warnings(warnings), err
	}

	return stack, Warnings(warnings), nil
}
