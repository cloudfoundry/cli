package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// Application represents a Cloud Controller V3 Application.
type Application struct {
	// GUID is the unique application identifier.
	GUID string
	// StackName is the name of the stack on which the application runs.
	StackName string
	// LifecycleBuildpacks is a list of the names of buildpacks.
	LifecycleBuildpacks []string
	// LifecycleType is the type of the lifecycle.
	LifecycleType constant.AppLifecycleType
	// Metadata is used for custom tagging of API resources
	Metadata *Metadata
	// Name is the name given to the application.
	Name string
	// Relationships list the relationships to the application.
	Relationships Relationships
	// State is the desired state of the application.
	State constant.ApplicationState
}

type ApplicationFeature struct {
	// Name of the application feature
	Name string

	Enabled bool
}

// MarshalJSON converts an Application into a Cloud Controller Application.
func (a Application) MarshalJSON() ([]byte, error) {
	ccApp := ccApplication{
		Name:          a.Name,
		Relationships: a.Relationships,
		Metadata:      a.Metadata,
	}

	if a.LifecycleType == constant.AppLifecycleTypeDocker {
		ccApp.setDockerLifecycle()
	} else if a.LifecycleType == constant.AppLifecycleTypeBuildpack {
		if len(a.LifecycleBuildpacks) > 0 || a.StackName != "" {
			if a.hasAutodetectedBuildpack() {
				ccApp.setAutodetectedBuildpackLifecycle(a)
			} else {
				ccApp.setBuildpackLifecycle(a)
			}
		}
	}

	return json.Marshal(ccApp)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Application response.
func (a *Application) UnmarshalJSON(data []byte) error {
	lifecycle := ccLifecycle{}
	ccApp := ccApplication{
		Lifecycle: &lifecycle,
	}

	err := cloudcontroller.DecodeJSON(data, &ccApp)
	if err != nil {
		return err
	}

	a.GUID = ccApp.GUID
	a.StackName = lifecycle.Data.Stack
	a.LifecycleBuildpacks = lifecycle.Data.Buildpacks
	a.LifecycleType = lifecycle.Type
	a.Name = ccApp.Name
	a.Relationships = ccApp.Relationships
	a.State = ccApp.State
	a.Metadata = ccApp.Metadata

	return nil
}

func (a Application) hasAutodetectedBuildpack() bool {
	if len(a.LifecycleBuildpacks) == 0 {
		return false
	}
	return a.LifecycleBuildpacks[0] == constant.AutodetectBuildpackValueDefault || a.LifecycleBuildpacks[0] == constant.AutodetectBuildpackValueNull
}

type ccLifecycle struct {
	Type constant.AppLifecycleType `json:"type,omitempty"`
	Data struct {
		Buildpacks []string `json:"buildpacks,omitempty"`
		Stack      string   `json:"stack,omitempty"`
	} `json:"data"`
}

type ccApplication struct {
	Name          string                    `json:"name,omitempty"`
	Relationships Relationships             `json:"relationships,omitempty"`
	Lifecycle     interface{}               `json:"lifecycle,omitempty"`
	GUID          string                    `json:"guid,omitempty"`
	State         constant.ApplicationState `json:"state,omitempty"`
	Metadata      *Metadata                 `json:"metadata,omitempty"`
}

func (ccApp *ccApplication) setAutodetectedBuildpackLifecycle(a Application) {
	var nullBuildpackLifecycle struct {
		Type constant.AppLifecycleType `json:"type,omitempty"`
		Data struct {
			Buildpacks []string `json:"buildpacks"`
			Stack      string   `json:"stack,omitempty"`
		} `json:"data"`
	}
	nullBuildpackLifecycle.Type = constant.AppLifecycleTypeBuildpack
	nullBuildpackLifecycle.Data.Stack = a.StackName
	ccApp.Lifecycle = nullBuildpackLifecycle
}

func (ccApp *ccApplication) setBuildpackLifecycle(a Application) {
	var lifecycle ccLifecycle
	lifecycle.Type = a.LifecycleType
	lifecycle.Data.Buildpacks = a.LifecycleBuildpacks
	lifecycle.Data.Stack = a.StackName
	ccApp.Lifecycle = lifecycle
}

func (ccApp *ccApplication) setDockerLifecycle() {
	ccApp.Lifecycle = ccLifecycle{
		Type: constant.AppLifecycleTypeDocker,
	}
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
		DecodeJSONResponseInto: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}

// GetApplications lists applications with optional queries.
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
		DecodeJSONResponseInto: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}

// UpdateApplicationRestart restarts the given application.
func (client *Client) UpdateApplicationRestart(appGUID string) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostApplicationActionRestartRequest,
		URIParams:   map[string]string{"app_guid": appGUID},
	})
	if err != nil {
		return Application{}, nil, err
	}

	var responseApp Application
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
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
		DecodeJSONResponseInto: &responseApp,
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
		DecodeJSONResponseInto: &responseApp,
	}
	err = client.connection.Make(request, &response)

	return responseApp, response.Warnings, err
}

// UpdateSSH enables/disables the ability to ssh for a given application.
func (client *Client) UpdateSSH(appGUID string, enabled bool) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PatchApplicationFeaturesRequest,
		Body:        bytes.NewReader([]byte(fmt.Sprintf(`{"enabled":%t}`, enabled))),
		URIParams:   map[string]string{"app_guid": appGUID, "name": "ssh"},
	})

	if err != nil {
		return nil, err
	}

	response := cloudcontroller.Response{
	}
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}

func (client *Client) GetSSH(appGUID string) (ApplicationFeature, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetApplicationFeaturesRequest,
		URIParams:   map[string]string{"app_guid": appGUID, "name": "ssh"},
	})

	if err != nil {
		return ApplicationFeature{}, nil, err
	}

	var applicationFeature ApplicationFeature
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &applicationFeature,
	}

	_ = client.connection.Make(request, &response)

	return applicationFeature, response.Warnings, nil
}
