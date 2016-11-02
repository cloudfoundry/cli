package ccv3

import (
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

type Application struct {
	Name string
	GUID string
}

func (client *CloudControllerClient) GetApplications(query url.Values) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		URI:    "/v3/apps",
		Method: http.MethodGet,
		Query:  query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullAppsList []Application
	var fullWarningsList Warnings

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
		request, err = client.newHTTPRequest(requestOptions{
			URI:    wrapper.NextURL,
			Method: http.MethodGet,
		})
		if err != nil {
			return nil, fullWarningsList, err
		}
	}

	return fullAppsList, fullWarningsList, nil
}
