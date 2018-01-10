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
	Name                string                    `json:"name,omitempty"`
	Relationships       Relationships             `json:"relationships,omitempty"`
	GUID                string                    `json:"guid,omitempty"`
	State               constant.ApplicationState `json:"state,omitempty"`
	LifecycleType       constant.AppLifecycleType `json:"-"`
	LifecycleBuildpacks []string                  `json:"-"`
}

func (a Application) MarshalJSON() ([]byte, error) {
	type rawApp Application
	var ccApp struct {
		Lifecycle map[string]interface{} `json:"lifecycle,omitempty"`

		rawApp
	}

	ccApp.rawApp = (rawApp)(a)

	switch a.LifecycleType {
	case constant.BuildpackAppLifecycleType:
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
	case constant.DockerAppLifecycleType:
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

	err := json.Unmarshal(data, &ccApp)
	if err != nil {
		return err
	}

	a.LifecycleType = ccApp.Lifecycle.Type
	a.LifecycleBuildpacks = ccApp.Lifecycle.Data.Buildpacks

	return nil
}

// DropletRelationship represents the relationship between a V3 Droplet and its
// V3 Application
type DropletRelationship struct {
	Relationship Relationship `json:"data"`
	Links        APILinks     `json:"links"`
}

// GetApplications lists applications with optional filters.
func (client *Client) GetApplications(query ...Query) ([]Application, Warnings, error) {
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

// SetApplicationDroplet sets the specified droplet on the given application.
func (client *Client) SetApplicationDroplet(appGUID string, dropletGUID string) (Relationship, Warnings, error) {
	relationship := Relationship{GUID: dropletGUID}
	bodyBytes, err := json.Marshal(relationship)
	if err != nil {
		return Relationship{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationCurrentDropletRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Relationship{}, nil, err
	}

	var responseRelationship Relationship
	response := cloudcontroller.Response{
		Result: &responseRelationship,
	}
	err = client.connection.Make(request, &response)

	return responseRelationship, response.Warnings, err
}

// StopApplication stops the given application.
func (client *Client) StopApplication(appGUID string) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationStopRequest,
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

// StartApplication starts the given application.
func (client *Client) StartApplication(appGUID string) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationStartRequest,
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
