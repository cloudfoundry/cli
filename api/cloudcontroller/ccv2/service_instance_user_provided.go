package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// GetUserProvidedServiceInstances returns back a list of *user provided* Service Instances based
// off the provided queries.
func (client *Client) GetUserProvidedServiceInstances(filters ...Filter) ([]ServiceInstance, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetUserProvidedServiceInstancesRequest,
		Query:       ConvertFilterParameters(filters),
	})
	if err != nil {
		return nil, nil, err
	}

	var fullInstancesList []ServiceInstance
	warnings, err := client.paginate(request, ServiceInstance{}, func(item interface{}) error {
		if instance, ok := item.(ServiceInstance); ok {
			fullInstancesList = append(fullInstancesList, instance)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServiceInstance{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullInstancesList, warnings, err
}

// UserProvidedServiceInstance represents the data to update a user-provided
// service instance. All fields are optional.
type UserProvidedServiceInstance struct {
	// Tags on the user-provided service instance
	Tags *[]string `json:"tags,omitempty"`
	// SyslogDrainURL for the user-provided service instance
	SyslogDrainURL *string `json:"syslog_drain_url,omitempty"`
	// RouteServiceURL for the user-provided service instance
	RouteServiceURL *string `json:"route_service_url,omitempty"`
	// Credentials for the user-provided service instance
	Credentials interface{} `json:"credentials,omitempty"`
}

func (u UserProvidedServiceInstance) WithTags(tags []string) UserProvidedServiceInstance {
	if tags == nil {
		tags = []string{}
	}

	u.Tags = &tags
	return u
}

func (u UserProvidedServiceInstance) WithSyslogDrainURL(url string) UserProvidedServiceInstance {
	u.SyslogDrainURL = &url
	return u
}

func (u UserProvidedServiceInstance) WithRouteServiceURL(url string) UserProvidedServiceInstance {
	u.RouteServiceURL = &url
	return u
}

func (u UserProvidedServiceInstance) WithCredentials(creds map[string]interface{}) UserProvidedServiceInstance {
	if creds == nil {
		creds = make(map[string]interface{})
	}

	u.Credentials = creds
	return u
}

func (client *Client) UpdateUserProvidedServiceInstance(serviceGUID string, instance UserProvidedServiceInstance) (Warnings, error) {
	bodyBytes, err := json.Marshal(instance)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutUserProvidedServiceInstance,
		URIParams:   Params{"user_provided_service_instance_guid": serviceGUID},
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
