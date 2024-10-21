package ccv2

import (
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv2/internal"
)

// Stack represents a Cloud Controller Stack.
type Stack struct {
	// GUID is the unique stack identifier.
	GUID string

	// Name is the name given to the stack.
	Name string

	// Description is the description of the stack.
	Description string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Stack response.
func (stack *Stack) UnmarshalJSON(data []byte) error {
	var ccStack struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccStack)
	if err != nil {
		return err
	}

	stack.GUID = ccStack.Metadata.GUID
	stack.Name = ccStack.Entity.Name
	stack.Description = ccStack.Entity.Description
	return nil
}

// GetStack returns the requested stack.
func (client *Client) GetStack(guid string) (Stack, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetStackRequest,
		URIParams:   Params{"stack_guid": guid},
	})
	if err != nil {
		return Stack{}, nil, err
	}

	var stack Stack
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &stack,
	}

	err = client.connection.Make(request, &response)
	return stack, response.Warnings, err
}

// GetStacks returns a list of Stacks based off of the provided filters.
func (client *Client) GetStacks(filters ...Filter) ([]Stack, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetStacksRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullStacksList []Stack
	warnings, err := client.paginate(request, Stack{}, func(item interface{}) error {
		if space, ok := item.(Stack); ok {
			fullStacksList = append(fullStacksList, space)
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
