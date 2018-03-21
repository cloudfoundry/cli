package ccv2

import (
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ApplicationInstance represents a Cloud Controller Application Instance.
type ApplicationInstance struct {
	// Details are arbitrary information about the instance.
	Details string

	// ID is the instance ID.
	ID int

	// Since is the Unix time stamp that represents the time the instance was
	// created.
	Since float64

	// State is the instance's state.
	State constant.ApplicationInstanceState
}

// UnmarshalJSON helps unmarshal a Cloud Controller application instance
// response.
func (instance *ApplicationInstance) UnmarshalJSON(data []byte) error {
	var ccInstance struct {
		Details string  `json:"details"`
		Since   float64 `json:"since"`
		State   string  `json:"state"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccInstance)
	if err != nil {
		return err
	}

	instance.Details = ccInstance.Details
	instance.State = constant.ApplicationInstanceState(ccInstance.State)
	instance.Since = ccInstance.Since

	return nil
}

// GetApplicationApplicationInstances returns a list of ApplicationInstance for
// a given application. Depending on the state of an application, it might skip
// some application instances.
func (client *Client) GetApplicationApplicationInstances(guid string) (map[int]ApplicationInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppInstancesRequest,
		URIParams:   Params{"app_guid": guid},
	})
	if err != nil {
		return nil, nil, err
	}

	var instances map[string]ApplicationInstance
	response := cloudcontroller.Response{
		Result: &instances,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return nil, response.Warnings, err
	}

	returnedInstances := map[int]ApplicationInstance{}
	for instanceID, instance := range instances {
		id, convertErr := strconv.Atoi(instanceID)
		if convertErr != nil {
			return nil, response.Warnings, convertErr
		}
		instance.ID = id
		returnedInstances[id] = instance
	}

	return returnedInstances, response.Warnings, nil
}
