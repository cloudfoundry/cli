package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServiceBinding represents a Cloud Controller Service Binding.
type ServiceBinding struct {
	GUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service Binding response.
func (serviceBinding *ServiceBinding) UnmarshalJSON(data []byte) error {
	var ccServiceBinding struct {
		Metadata internal.Metadata
	}
	err := json.Unmarshal(data, &ccServiceBinding)
	if err != nil {
		return err
	}

	serviceBinding.GUID = ccServiceBinding.Metadata.GUID
	return nil
}

// GetServiceBindings returns back a list of Service Bindings based off of the
// provided queries.
func (client *CloudControllerClient) GetServiceBindings(queries []Query) ([]ServiceBinding, Warnings, error) {
	request := cloudcontroller.NewRequest(
		internal.ServiceBindingsRequest,
		nil,
		nil,
		FormatQueryParameters(queries),
	)

	allServiceBindingsList := []ServiceBinding{}
	allWarningsList := Warnings{}

	for {
		var serviceBindings []ServiceBinding
		wrapper := PaginatedWrapper{
			Resources: &serviceBindings,
		}
		response := cloudcontroller.Response{
			Result: &wrapper,
		}

		err := client.connection.Make(request, &response)
		allWarningsList = append(allWarningsList, response.Warnings...)
		if err != nil {
			return nil, allWarningsList, err
		}

		allServiceBindingsList = append(allServiceBindingsList, serviceBindings...)

		if wrapper.NextURL == "" {
			break
		}
		request = cloudcontroller.NewRequestFromURI(
			wrapper.NextURL,
			"GET",
			nil,
		)
	}

	return allServiceBindingsList, allWarningsList, nil
}

// DeleteServiceBinding will destroy the requested Service Binding.
func (client *CloudControllerClient) DeleteServiceBinding(serviceBindingGUID string) (Warnings, error) {
	request := cloudcontroller.NewRequest(
		internal.DeleteServiceBindingRequest,
		map[string]string{"service_binding_guid": serviceBindingGUID},
		nil,
		nil,
	)

	var response cloudcontroller.Response
	err := client.connection.Make(request, &response)
	return response.Warnings, err
}
