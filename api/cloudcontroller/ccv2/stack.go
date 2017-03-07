package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// Stack represents a Cloud Controller Stack.
type Stack struct {
	GUID        string
	Name        string
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
	if err := json.Unmarshal(data, &ccStack); err != nil {
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
		Result: &stack,
	}

	err = client.connection.Make(request, &response)
	return stack, response.Warnings, err
}
