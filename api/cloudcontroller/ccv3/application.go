package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// CreateApplication creates an application with the given settings.
func (client *Client) CreateApplication(app resources.Application) (resources.Application, Warnings, error) {
	var responseBody resources.Application

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationRequest,
		RequestBody:  app,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

func (client *Client) GetApplicationByNameAndSpace(appName string, spaceGUID string) (resources.Application, Warnings, error) {
	apps, warnings, err := client.GetApplications(
		Query{Key: NameFilter, Values: []string{appName}},
		Query{Key: SpaceGUIDFilter, Values: []string{spaceGUID}},
	)
	if err != nil {
		return resources.Application{}, warnings, err
	}

	if len(apps) == 0 {
		return resources.Application{}, warnings, ccerror.ApplicationNotFoundError{Name: appName}
	}

	return apps[0], warnings, nil
}

// GetApplications lists applications with optional queries.
func (client *Client) GetApplications(query ...Query) ([]resources.Application, Warnings, error) {
	var apps []resources.Application

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetApplicationsRequest,
		Query:        query,
		ResponseBody: resources.Application{},
		AppendToList: func(item interface{}) error {
			apps = append(apps, item.(resources.Application))
			return nil
		},
	})

	return apps, warnings, err
}

// UpdateApplication updates an application with the given settings.
func (client *Client) UpdateApplication(app resources.Application) (resources.Application, Warnings, error) {
	var responseBody resources.Application

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchApplicationRequest,
		URIParams:    internal.Params{"app_guid": app.GUID},
		RequestBody:  app,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateApplicationRestart restarts the given application.
func (client *Client) UpdateApplicationRestart(appGUID string) (resources.Application, Warnings, error) {
	var responseBody resources.Application

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationActionRestartRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateApplicationStart starts the given application.
func (client *Client) UpdateApplicationStart(appGUID string) (resources.Application, Warnings, error) {
	var responseBody resources.Application

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationActionStartRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateApplicationStop stops the given application.
func (client *Client) UpdateApplicationStop(appGUID string) (resources.Application, Warnings, error) {
	var responseBody resources.Application

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostApplicationActionStopRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
