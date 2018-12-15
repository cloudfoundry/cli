package ccv2

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ServicePlan represents a predefined set of configurations for a Cloud
// Controller service object.
type ServicePlan struct {
	//GUID is the unique identifier of the service plan.
	GUID string

	// Name is the name of the service plan.
	Name string

	// ServiceGUID is the unique identifier of the service that the service
	// plan belongs to.
	ServiceGUID string

	// Public is true if plan is accessible to all organizations.
	Public bool

	// Description of the plan
	Description string

	// Free is true if plan is free
	Free bool
}

// UnmarshalJSON helps unmarshal a Cloud Controller Service Plan response.
func (servicePlan *ServicePlan) UnmarshalJSON(data []byte) error {
	var ccServicePlan struct {
		Metadata internal.Metadata
		Entity   struct {
			Name        string `json:"name"`
			ServiceGUID string `json:"service_guid"`
			Public      bool   `json:"public"`
			Description string `json:"description"`
			Free        bool   `json:"free"`
		}
	}
	err := cloudcontroller.DecodeJSON(data, &ccServicePlan)
	if err != nil {
		return err
	}

	servicePlan.GUID = ccServicePlan.Metadata.GUID
	servicePlan.Name = ccServicePlan.Entity.Name
	servicePlan.ServiceGUID = ccServicePlan.Entity.ServiceGUID
	servicePlan.Public = ccServicePlan.Entity.Public
	servicePlan.Description = ccServicePlan.Entity.Description
	servicePlan.Free = ccServicePlan.Entity.Free
	return nil
}

// GetServicePlan returns the service plan with the given GUID.
func (client *Client) GetServicePlan(servicePlanGUID string) (ServicePlan, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServicePlanRequest,
		URIParams:   Params{"service_plan_guid": servicePlanGUID},
	})
	if err != nil {
		return ServicePlan{}, nil, err
	}

	var servicePlan ServicePlan
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &servicePlan,
	}

	err = client.connection.Make(request, &response)
	return servicePlan, response.Warnings, err
}

func (client *Client) GetServicePlans(filters ...Filter) ([]ServicePlan, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetServicePlansRequest,
		Query:       ConvertFilterParameters(filters),
	})

	if err != nil {
		return nil, nil, err
	}

	var fullServicePlansList []ServicePlan
	warnings, err := client.paginate(request, ServicePlan{}, func(item interface{}) error {
		if plan, ok := item.(ServicePlan); ok {
			fullServicePlansList = append(fullServicePlansList, plan)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   ServicePlan{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullServicePlansList, warnings, err
}

type updateServicePlanRequestBody struct {
	Public bool `json:"public"`
}

func (client *Client) UpdateServicePlan(guid string, public bool) (Warnings, error) {
	requestBody := updateServicePlanRequestBody{
		Public: public,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutServicePlanRequest,
		Body:        bytes.NewReader(body),
		URIParams:   Params{"service_plan_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
