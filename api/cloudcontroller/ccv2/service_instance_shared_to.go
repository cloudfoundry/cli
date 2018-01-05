package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type ServiceInstanceSharedTo struct {
	SpaceGUID        string `json:"space_guid"`
	SpaceName        string `json:"space_name"`
	OrganizationName string `json:"organization_name"`
	BoundAppCount    int    `json:"bound_app_count"`
}

func (client *Client) GetServiceInstanceSharedTos(serviceInstanceGUID string) ([]ServiceInstanceSharedTo, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServiceInstanceSharedToRequest,
		URIParams:   Params{"service_instance_guid": serviceInstanceGUID},
	})

	if err != nil {
		return nil, nil, err
	}

	var fullSharedToList []ServiceInstanceSharedTo
	warnings, err := client.paginate(request, ServiceInstanceSharedTo{}, func(item interface{}) error {
		if instance, ok := item.(ServiceInstanceSharedTo); ok {
			fullSharedToList = append(fullSharedToList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstanceSharedTo{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSharedToList, warnings, err
}
