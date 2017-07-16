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
	Name          string
	Relationships Relationships
	GUID          string
	State         string
	Buildpacks    []string
}

func (a Application) MarshalJSON() ([]byte, error) {
	var ccApp struct {
		Name          string                 `json:"name,omitempty"`
		Relationships Relationships          `json:"relationships,omitempty"`
		Lifecycle     map[string]interface{} `json:"lifecycle,omitempty"`
	}

	ccApp.Name = a.Name
	ccApp.Relationships = a.Relationships
	if len(a.Buildpacks) > 0 {
		switch a.Buildpacks[0] {
		case "default", "null":
			ccApp.Lifecycle = map[string]interface{}{
				"type": "buildpack",
				"data": map[string]interface{}{
					"buildpacks": nil,
				},
			}
		default:
			ccApp.Lifecycle = map[string]interface{}{
				"type": "buildpack",
				"data": map[string]interface{}{
					"buildpacks": a.Buildpacks,
				},
			}
		}
	}

	return json.Marshal(ccApp)
}

// UnmarshalJSON helps unmarshal a Cloud Controller V3 Application response
func (a *Application) UnmarshalJSON(data []byte) error {
	// TODO: do we care about rebuilding the Relationships object?
	var ccApp struct {
		Name      string `json:"name"`
		GUID      string `json:"guid"`
		State     string `json:"state,omitempty"`
		Lifecycle struct {
			Type string `json:"type"`
			Data struct {
				Buildpacks []string `json:"buildpacks"`
			} `json:"data"`
		} `json:"lifecycle,omitempty"`
	}

	if err := json.Unmarshal(data, &ccApp); err != nil {
		return err
	}

	a.Name = ccApp.Name
	a.GUID = ccApp.GUID
	a.State = ccApp.State
	a.Buildpacks = ccApp.Lifecycle.Data.Buildpacks

	return nil
}

// DropletRelationship represents the relationship between a V3 Droplet and its
// V3 Application
type DropletRelationship struct {
	Relationship Relationship `json:"data"`
	Links        APILinks     `json:"links"`
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

// UpdateApplication updates an application with the given settings
func (client *Client) UpdateApplication(app Application) (Application, Warnings, error) {
	bodyBytes, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationRequest,
		Body:        bytes.NewReader(bodyBytes),
		URIParams:   map[string]string{"guid": app.GUID},
	})

	var responseApp Application
	response := cloudcontroller.Response{
		Result: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}

func (client *Client) SetApplicationDroplet(appGUID string, dropletGUID string) (Relationship, Warnings, error) {
	relationship := Relationship{GUID: dropletGUID}
	bodyBytes, err := json.Marshal(relationship)
	if err != nil {
		return Relationship{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationCurrentDropletRequest,
		URIParams:   map[string]string{"guid": appGUID},
		Body:        bytes.NewReader(bodyBytes),
	})

	var responseRelationship Relationship
	response := cloudcontroller.Response{
		Result: &responseRelationship,
	}
	err = client.connection.Make(request, &response)

	return responseRelationship, response.Warnings, err
}

func (client *Client) StopApplication(appGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutApplicationStopRequest,
		URIParams:   map[string]string{"guid": appGUID},
	})

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

func (client *Client) StartApplication(appGUID string) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutApplicationStartRequest,
		URIParams:   map[string]string{"guid": appGUID},
	})

	var responseApp Application
	response := cloudcontroller.Response{
		Result: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}
