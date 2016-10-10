package ccv2

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type Application struct {
	GUID string
	Name string
}

func (application *Application) UnmarshalJSON(data []byte) error {
	var ccApp struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Name string `json:"name"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(data, &ccApp); err != nil {
		return err
	}

	application.GUID = ccApp.Metadata.GUID
	application.Name = ccApp.Entity.Name
	return nil
}

func (client *CloudControllerClient) GetApplications(queryParams []Query) ([]Application, Warnings, error) {
	request := cloudcontroller.Request{
		RequestName: internal.AppsRequest,
		Query:       FormatQueryParameters(queryParams),
	}

	fullAppsList := []Application{}
	fullWarningsList := Warnings{}

	for {
		var apps []Application
		wrapper := PaginatedWrapper{
			Resources: &apps,
		}
		response := cloudcontroller.Response{
			Result: &wrapper,
		}

		err := client.connection.Make(request, &response)
		fullWarningsList = append(fullWarningsList, response.Warnings...)
		if err != nil {
			return nil, fullWarningsList, err
		}
		fullAppsList = append(fullAppsList, apps...)

		if wrapper.NextURL == "" {
			break
		}
		request = cloudcontroller.Request{
			URI:    wrapper.NextURL,
			Method: "GET",
		}
	}

	return fullAppsList, fullWarningsList, nil
}
