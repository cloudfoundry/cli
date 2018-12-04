package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Stack struct {
	// GUID is a unique space identifier.
	GUID string `json:"guid"`
	// Name is the name of the space.
	Name string `json:"name"`
	// Description is the description for the stack
	Description string `json:"description"`
}

// GetSpaces lists spaces with optional filters.
func (client *Client) GetStacks(query ...Query) ([]Stack, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetStacksRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullStacksList []Stack
	warnings, err := client.paginate(request, Stack{}, func(item interface{}) error {
		if stack, ok := item.(Stack); ok {
			fullStacksList = append(fullStacksList, stack)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Stack{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullStacksList, warnings, err
}
