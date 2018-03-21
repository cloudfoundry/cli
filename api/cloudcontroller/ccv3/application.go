package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Application represents a Cloud Controller V3 Application.
type Application struct {
	// GUID is the unique application identifier.
	GUID string `json:"guid,omitempty"`
	// LifecycleBuildpacks is a list of the names of buildpacks.
	LifecycleBuildpacks []string `json:"-"`
	// LifecycleType is the type of the lifecycle.
	LifecycleType constant.AppLifecycleType `json:"-"`
	// Name is the name given to the application.
	Name string `json:"name,omitempty"`
	// Relationships list the relationships to the application.
	Relationships Relationships `json:"relationships,omitempty"`
	// State is the desired state of the application.
	State constant.ApplicationState `json:"state,omitempty"`
}

func (a Application) MarshalJSON() ([]byte, error) {
	type rawApp Application
	var ccApp struct {
		Lifecycle map[string]interface{} `json:"lifecycle,omitempty"`

		rawApp
	}

	ccApp.rawApp = (rawApp)(a)

	switch a.LifecycleType {
	case constant.AppLifecycleTypeBuildpack:
		if len(a.LifecycleBuildpacks) > 0 {
			ccApp.Lifecycle = map[string]interface{}{}
			ccApp.Lifecycle["type"] = a.LifecycleType
			ccApp.Lifecycle["data"] = map[string]interface{}{}
			switch a.LifecycleBuildpacks[0] {
			case "default", "null":
				ccApp.Lifecycle["data"] = map[string][]string{
					"buildpacks": nil,
				}
			default:
				ccApp.Lifecycle["data"] = map[string][]string{
					"buildpacks": a.LifecycleBuildpacks,
				}
			}
		}
	case constant.AppLifecycleTypeDocker:
		ccApp.Lifecycle = map[string]interface{}{}
		ccApp.Lifecycle["type"] = a.LifecycleType
		ccApp.Lifecycle["data"] = map[string]interface{}{}
	}

	ccApp.GUID = ""
	return json.Marshal(ccApp)
}

func (a *Application) UnmarshalJSON(data []byte) error {
	type rawApp Application
	var ccApp struct {
		*rawApp

		Lifecycle struct {
			Type constant.AppLifecycleType
			Data struct {
				Buildpacks []string
			}
		}
	}

	ccApp.rawApp = (*rawApp)(a)

	err := cloudcontroller.DecodeJSON(data, &ccApp)
	if err != nil {
		return err
	}

	a.LifecycleType = ccApp.Lifecycle.Type
	a.LifecycleBuildpacks = ccApp.Lifecycle.Data.Buildpacks

	return nil
}

// CreateApplication creates an application with the given settings.
func (client *Client) CreateApplication(app Application) (Application, Warnings, error) {
	bodyBytes, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Application{}, nil, err
	}

	var responseApp Application
	response := cloudcontroller.Response{
		Result: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}

// DeleteApplication deletes the app with the given app GUID.
func (client *Client) DeleteApplication(appGUID string) (string, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteApplicationRequest,
		URIParams:   internal.Params{"app_guid": appGUID},
	})
	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.ResourceLocationURL, response.Warnings, err
}

// GetApplications lists applications with optional filters.
func (client *Client) GetApplications(query ...Query) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationsRequest,
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

// UpdateApplication updates an application with the given settings.
func (client *Client) UpdateApplication(app Application) (Application, Warnings, error) {
	bodyBytes, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationRequest,
		Body:        bytes.NewReader(bodyBytes),
		URIParams:   map[string]string{"app_guid": app.GUID},
	})
	if err != nil {
		return Application{}, nil, err
	}

	var responseApp Application
	response := cloudcontroller.Response{
		Result: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}

// UpdateApplicationApplyManifest applies the manifest to the given
// application.
func (client *Client) UpdateApplicationApplyManifest(appGUID string, rawManifest []byte) (string, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationActionApplyManifest,
		URIParams:   map[string]string{"app_guid": appGUID},
		Body:        bytes.NewReader(rawManifest),
	})

	if err != nil {
		return "", nil, err
	}

	request.Header.Set("Content-Type", "application/x-yaml")

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.ResourceLocationURL, response.Warnings, err
}

// UpdateApplicationStart starts the given application.
func (client *Client) UpdateApplicationStart(appGUID string) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationActionStartRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
	})
	if err != nil {
		return Application{}, nil, err
	}

	var responseApp Application
	response := cloudcontroller.Response{
		Result: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}

// UpdateApplicationStop stops the given application.
func (client *Client) UpdateApplicationStop(appGUID string) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationActionStopRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
	})
	if err != nil {
		return Application{}, nil, err
	}

	var responseApp Application
	response := cloudcontroller.Response{
		Result: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}
