package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type ServiceBinding struct {
	GUID string
}

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

func (client *CloudControllerClient) GetServiceBindings(queries []Query) ([]ServiceBinding, Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: ServiceBindingsRequest,
		Query:       FormatQueryParameters(queries),
	}

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
		request = cloudcontroller.Request{
			URI:    wrapper.NextURL,
			Method: "GET",
		}
	}

	return allServiceBindingsList, allWarningsList, nil
}

func (client *CloudControllerClient) DeleteServiceBinding(serviceBindingGUID string) (Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: DeleteServiceBindingRequest,
		Params:      map[string]string{"service_binding_guid": serviceBindingGUID},
	}

	var response cloudcontroller.Response
	err := client.connection.Make(request, &response)
	return response.Warnings, err
}
