package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServicePlanVisibility represents a Cloud Controller Service Plan Visibility.
type ServicePlanVisibility struct {
	// GUID is the unique Service Plan Visibility identifier.
	GUID string
	// ServicePlanGUID of the associated Service Plan.
	ServicePlanGUID string
	// OrganizationGUID of the associated Organization.
	OrganizationGUID string
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service Plan Visibilities
// response.
func (servicePlanVisibility *ServicePlanVisibility) UnmarshalJSON(data []byte) error {
	var ccServicePlanVisibility struct {
		Metadata internal.Metadata
		Entity   struct {
			ServicePlanGUID  string `json:"service_plan_guid"`
			OrganizationGUID string `json:"organization_guid"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccServicePlanVisibility)
	if err != nil {
		return err
	}

	servicePlanVisibility.GUID = ccServicePlanVisibility.Metadata.GUID
	servicePlanVisibility.ServicePlanGUID = ccServicePlanVisibility.Entity.ServicePlanGUID
	servicePlanVisibility.OrganizationGUID = ccServicePlanVisibility.Entity.OrganizationGUID
	return nil
}

type createServicePlanRequestBody struct {
	ServicePlanGUID  string `json:"service_plan_guid"`
	OrganizationGUID string `json:"organization_guid"`
}

func (client *Client) CreateServicePlanVisibility(planGUID string, orgGUID string) (ServicePlanVisibility, Warnings, error) {
	requestBody := createServicePlanRequestBody{
		ServicePlanGUID:  planGUID,
		OrganizationGUID: orgGUID,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return ServicePlanVisibility{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostServicePlanVisibilityRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return ServicePlanVisibility{}, nil, err
	}

	var servicePlanVisibility ServicePlanVisibility
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &servicePlanVisibility,
	}

	err = client.connection.Make(request, &response)
	return servicePlanVisibility, response.Warnings, err
}

func (client *Client) DeleteServicePlanVisibility(servicePlanVisibilityGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteServicePlanVisibilityRequest,
		URIParams:   Params{"service_plan_visibility_guid": servicePlanVisibilityGUID},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}

	err = client.connection.Make(request, &response)
	return response.Warnings, err
}

// GetServicePlanVisibilities returns back a list of Service Plan Visibilities
// given the provided filters.
func (client *Client) GetServicePlanVisibilities(filters ...Filter) ([]ServicePlanVisibility, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServicePlanVisibilitiesRequest,
		Query:       ConvertFilterParameters(filters),
	})

	if err != nil {
		return nil, nil, err
	}

	var fullVisibilityList []ServicePlanVisibility
	warnings, err := client.paginate(request, ServicePlanVisibility{}, func(item interface{}) error {
		if vis, ok := item.(ServicePlanVisibility); ok {
			fullVisibilityList = append(fullVisibilityList, vis)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServicePlanVisibility{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullVisibilityList, warnings, err
}
