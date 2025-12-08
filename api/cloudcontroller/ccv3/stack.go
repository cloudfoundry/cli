package ccv3

import (
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v9/resources"
)

// GetStacks lists stacks with optional filters.
func (client *Client) GetStacks(query ...Query) ([]resources.Stack, Warnings, error) {
	var stacks []resources.Stack

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetStacksRequest,
		Query:        query,
		ResponseBody: resources.Stack{},
		AppendToList: func(item interface{}) error {
			stacks = append(stacks, item.(resources.Stack))
			return nil
		},
	})

	return stacks, warnings, err
}

// UpdateStack updates a stack's state.
func (client *Client) UpdateStack(stackGUID string, state string) (resources.Stack, Warnings, error) {
	var responseStack resources.Stack

	type StackUpdate struct {
		State string `json:"state"`
	}

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchStackRequest,
		URIParams:    internal.Params{"stack_guid": stackGUID},
		RequestBody:  StackUpdate{State: state},
		ResponseBody: &responseStack,
	})

	return responseStack, warnings, err
}
