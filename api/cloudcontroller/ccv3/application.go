package ccv3

import (
	"bytes"
	"encoding/json"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Application represents a Cloud Controller V3 Application.
type Application struct {
	Name          string        `json:"name"`
	GUID          string        `json:"guid,omitempty"`
	Relationships Relationships `json:"relationships"`
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
	warnings, err := client.paginate(request, Application{}, func(item interface{}) error {
		if app, ok := item.(Application); ok {
			fullAppsList = append(fullAppsList, app)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Application{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullAppsList, warnings, err
}

// CreateApplication creates an application with the given settings
func (client *Client) CreateApplication(app Application) (Application, Warnings, error) {
	bodyBytes, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationRequest,
		Body:        bytes.NewReader(bodyBytes),
	})

	var responseApp Application
	response := cloudcontroller.Response{
		Result: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}
