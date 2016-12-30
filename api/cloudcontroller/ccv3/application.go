package ccv3

import (
	"net/http"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Application represents a Cloud Controller V3 Application.
type Application struct {
	Name string
	GUID string
}

// GetApplications lists applications with optional filters.
func (client *Client) GetApplications(query url.Values) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppsRequest,
		Query:       query,
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

		if wrapper.Pagination.Next.HREF == "" {
			break
		}
		request, err = client.newHTTPRequest(requestOptions{
			URL:    wrapper.Pagination.Next.HREF,
			Method: http.MethodGet,
		})
		if err != nil {
			return nil, fullWarningsList, err
		}
	}

	return fullAppsList, fullWarningsList, nil
}
