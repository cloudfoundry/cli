package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Stack struct {
	// GUID is a unique stack identifier.
	GUID string `json:"guid"`
	// Name is the name of the stack.
	Name string `json:"name"`
	// Description is the description for the stack
	Description string `json:"description"`

	// Metadata is used for custom tagging of API resources
	Metadata *Metadata `json:"metadata,omitempty"`
}

// GetStacks lists stacks with optional filters.
func (client *Client) GetStacks(query ...Query) ([]Stack, Warnings, error) {
	var resources []Stack

	_, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetStacksRequest,
		Query:        query,
		ResponseBody: Stack{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Stack))
			return nil
		},
	})

	return resources, warnings, err
}
