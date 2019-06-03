package ccv2

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// SpaceSummaryApplication represents an application inside a space
type SpaceSummaryApplication struct {
	Name         string   `json:"name"`
	ServiceNames []string `json:"service_names"`
}

// SpaceSummaryService represents a service inside a space
type SpaceSummaryService struct {
	GUID              string
	Label             string
	ServiceBrokerName string
}

// SpaceSummaryServicePlan represents a service plan inside a space
type SpaceSummaryServicePlan struct {
	GUID            string              `json:"guid"`
	Name            string              `json:"name"`
	Service         SpaceSummaryService `json:"service"`
	MaintenanceInfo MaintenanceInfo     `json:"maintenance_info"`
}

// SpaceSummaryServiceInstance represents a service instance inside a space
type SpaceSummaryServiceInstance struct {
	LastOperation   LastOperation           `json:"last_operation"`
	Name            string                  `json:"name"`
	ServicePlan     SpaceSummaryServicePlan `json:"service_plan"`
	MaintenanceInfo MaintenanceInfo         `json:"maintenance_info"`
}

// SpaceSummary represents a service instance inside a space
type SpaceSummary struct {
	Applications     []SpaceSummaryApplication
	Name             string
	ServiceInstances []SpaceSummaryServiceInstance
}

// UnmarshalJSON helps unmarshal a Cloud Controller Space Summary response.
func (spaceSummary *SpaceSummary) UnmarshalJSON(data []byte) error {
	var ccSpaceSummary struct {
		Applications     []SpaceSummaryApplication     `json:"apps"`
		Name             string                        `json:"name"`
		ServiceInstances []SpaceSummaryServiceInstance `json:"services"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccSpaceSummary)
	if err != nil {
		return err
	}

	spaceSummary.Name = ccSpaceSummary.Name
	spaceSummary.Applications = ccSpaceSummary.Applications
	spaceSummary.ServiceInstances = ccSpaceSummary.ServiceInstances

	return nil
}

// GetSpaceSummary returns the summary of the space with the given GUID.
func (client *Client) GetSpaceSummary(spaceGUID string) (SpaceSummary, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetSpaceSummaryRequest,
		URIParams:   Params{"space_guid": spaceGUID},
	})
	if err != nil {
		return SpaceSummary{}, nil, err
	}

	var spaceSummary SpaceSummary
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &spaceSummary,
	}

	err = client.connection.Make(request, &response)
	return spaceSummary, response.Warnings, err
}
