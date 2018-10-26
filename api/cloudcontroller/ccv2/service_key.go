package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceKey represents a Cloud Controller Service Key.
type ServiceKey struct {
	// GUID is the unique Service Key identifier.
	GUID string
	// Name is the name of the service key.
	Name string
	// ServiceInstanceGUID is the associated service instance GUID.
	ServiceInstanceGUID string
	// Credentials are the credentials returned by the service broker for the service key.
	Credentials map[string]interface{}
}

func (serviceKey *ServiceKey) UnmarshalJSON(data []byte) error {
	var ccServiceKey struct {
		Metadata internal.Metadata
		Entity   struct {
			ServiceInstanceGUID string                 `json:"service_instance_guid"`
			Name                string                 `json:"name"`
			Credentials         map[string]interface{} `json:"credentials"`
		}
	}
	err := cloudcontroller.DecodeJSON(data, &ccServiceKey)

	if err != nil {
		return err
	}

	serviceKey.GUID = ccServiceKey.Metadata.GUID
	serviceKey.Name = ccServiceKey.Entity.Name
	serviceKey.ServiceInstanceGUID = ccServiceKey.Entity.ServiceInstanceGUID
	serviceKey.Credentials = ccServiceKey.Entity.Credentials

	return nil
}

// serviceKeyRequestBody represents the body of the service key create
// request.
type serviceKeyRequestBody struct {
	ServiceInstanceGUID string                 `json:"service_instance_guid"`
	Name                string                 `json:"name"`
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
}

// CreateServiceKey creates a new service key using the provided name and
// parameters for the requested service instance.
func (client *Client) CreateServiceKey(serviceInstanceGUID string, keyName string, parameters map[string]interface{}) (ServiceKey, Warnings, error) {
	requestBody := serviceKeyRequestBody{
		ServiceInstanceGUID: serviceInstanceGUID,
		Name:                keyName,
		Parameters:          parameters,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return ServiceKey{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostServiceKeyRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return ServiceKey{}, nil, err
	}

	var serviceKey ServiceKey
	response := cloudcontroller.Response{
		Result: &serviceKey,
	}
	err = client.connection.Make(request, &response)

	return serviceKey, response.Warnings, err
}
