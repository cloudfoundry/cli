package ccv3

import (
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v7/resources"
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
