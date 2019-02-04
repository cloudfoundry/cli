package ccv2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
	"code.cloudfoundry.org/cli/types"
)

// Application represents a Cloud Controller Application.
type Application struct {
	// Buildpack is the buildpack set by the user.
	Buildpack types.FilteredString

	// Command is the user specified start command.
	Command types.FilteredString

	// DetectedBuildpack is the buildpack automatically detected.
	DetectedBuildpack types.FilteredString

	// DetectedStartCommand is the command used to start the application.
	DetectedStartCommand types.FilteredString

	// DiskQuota is the disk given to each instance, in megabytes.
	DiskQuota types.NullByteSizeInMb

	// DockerCredentials is the authentication information for the provided
	// DockerImage.
	DockerCredentials DockerCredentials

	// DockerImage is the docker image location.
	DockerImage string

	// EnvironmentVariables are the environment variables passed to the app.
	EnvironmentVariables map[string]string

	// GUID is the unique application identifier.
	GUID string

	// HealthCheckTimeout is the number of seconds for health checking of an
	// staged app when starting up.
	HealthCheckTimeout uint64

	// HealthCheckType is the type of health check that will be done to the app.
	HealthCheckType constant.ApplicationHealthCheckType

	// HealthCheckHTTPEndpoint is the url of the http health check endpoint.
	HealthCheckHTTPEndpoint string

	// Instances is the total number of app instances.
	Instances types.NullInt

	// Memory is the memory given to each instance, in megabytes.
	Memory types.NullByteSizeInMb

	// Name is the name given to the application.
	Name string

	// PackageState represents the staging state of the application bits.
	PackageState constant.ApplicationPackageState

	// PackageUpdatedAt is the last time the app bits were updated. In RFC3339.
	PackageUpdatedAt time.Time

	// SpaceGUID is the GUID of the app's space.
	SpaceGUID string

	// StackGUID is the GUID for the Stack the application is running on.
	StackGUID string

	// StagingFailedDescription is the verbose description of why the package
	// failed to stage.
	StagingFailedDescription string

	// StagingFailedReason is the reason why the package failed to stage.
	StagingFailedReason string

	// State is the desired state of the application.
	State constant.ApplicationState
}

// MarshalJSON converts an application into a Cloud Controller Application.
func (application Application) MarshalJSON() ([]byte, error) {
	ccApp := struct {
		Buildpack               *string                             `json:"buildpack,omitempty"`
		Command                 *string                             `json:"command,omitempty"`
		DiskQuota               *uint64                             `json:"disk_quota,omitempty"`
		DockerCredentials       *DockerCredentials                  `json:"docker_credentials,omitempty"`
		DockerImage             string                              `json:"docker_image,omitempty"`
		EnvironmentVariables    map[string]string                   `json:"environment_json,omitempty"`
		HealthCheckHTTPEndpoint *string                             `json:"health_check_http_endpoint,omitempty"`
		HealthCheckTimeout      uint64                              `json:"health_check_timeout,omitempty"`
		HealthCheckType         constant.ApplicationHealthCheckType `json:"health_check_type,omitempty"`
		Instances               *int                                `json:"instances,omitempty"`
		Memory                  *uint64                             `json:"memory,omitempty"`
		Name                    string                              `json:"name,omitempty"`
		SpaceGUID               string                              `json:"space_guid,omitempty"`
		StackGUID               string                              `json:"stack_guid,omitempty"`
		State                   constant.ApplicationState           `json:"state,omitempty"`
	}{
		DockerImage:          application.DockerImage,
		EnvironmentVariables: application.EnvironmentVariables,
		HealthCheckTimeout:   application.HealthCheckTimeout,
		HealthCheckType:      application.HealthCheckType,
		Name:                 application.Name,
		SpaceGUID:            application.SpaceGUID,
		StackGUID:            application.StackGUID,
		State:                application.State,
	}

	if application.Buildpack.IsSet {
		ccApp.Buildpack = &application.Buildpack.Value
	}

	if application.Command.IsSet {
		ccApp.Command = &application.Command.Value
	}

	if application.DiskQuota.IsSet {
		ccApp.DiskQuota = &application.DiskQuota.Value
	}

	if application.DockerCredentials.Username != "" || application.DockerCredentials.Password != "" {
		ccApp.DockerCredentials = &DockerCredentials{
			Username: application.DockerCredentials.Username,
			Password: application.DockerCredentials.Password,
		}
	}

	if application.Instances.IsSet {
		ccApp.Instances = &application.Instances.Value
	}

	if application.HealthCheckType != "" {
		ccApp.HealthCheckHTTPEndpoint = &application.HealthCheckHTTPEndpoint
	}

	if application.Memory.IsSet {
		ccApp.Memory = &application.Memory.Value
	}

	return json.Marshal(ccApp)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Application response.
func (application *Application) UnmarshalJSON(data []byte) error {
	var ccApp struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Buildpack            string            `json:"buildpack"`
			Command              string            `json:"command"`
			DetectedBuildpack    string            `json:"detected_buildpack"`
			DetectedStartCommand string            `json:"detected_start_command"`
			DiskQuota            *uint64           `json:"disk_quota"`
			DockerImage          string            `json:"docker_image"`
			DockerCredentials    DockerCredentials `json:"docker_credentials"`
			// EnvironmentVariables' values can be any type, so we must accept
			// interface{}, but we convert to string.
			EnvironmentVariables     map[string]interface{} `json:"environment_json"`
			HealthCheckHTTPEndpoint  string                 `json:"health_check_http_endpoint"`
			HealthCheckTimeout       uint64                 `json:"health_check_timeout"`
			HealthCheckType          string                 `json:"health_check_type"`
			Instances                json.Number            `json:"instances"`
			Memory                   *uint64                `json:"memory"`
			Name                     string                 `json:"name"`
			PackageState             string                 `json:"package_state"`
			PackageUpdatedAt         *time.Time             `json:"package_updated_at"`
			SpaceGUID                string                 `json:"space_guid"`
			StackGUID                string                 `json:"stack_guid"`
			StagingFailedDescription string                 `json:"staging_failed_description"`
			StagingFailedReason      string                 `json:"staging_failed_reason"`
			State                    string                 `json:"state"`
		} `json:"entity"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccApp)
	if err != nil {
		return err
	}

	application.Buildpack.ParseValue(ccApp.Entity.Buildpack)
	application.Command.ParseValue(ccApp.Entity.Command)
	application.DetectedBuildpack.ParseValue(ccApp.Entity.DetectedBuildpack)
	application.DetectedStartCommand.ParseValue(ccApp.Entity.DetectedStartCommand)
	application.DiskQuota.ParseUint64Value(ccApp.Entity.DiskQuota)
	application.DockerCredentials = ccApp.Entity.DockerCredentials
	application.DockerImage = ccApp.Entity.DockerImage
	application.GUID = ccApp.Metadata.GUID
	application.HealthCheckHTTPEndpoint = ccApp.Entity.HealthCheckHTTPEndpoint
	application.HealthCheckTimeout = ccApp.Entity.HealthCheckTimeout
	application.HealthCheckType = constant.ApplicationHealthCheckType(ccApp.Entity.HealthCheckType)
	application.Memory.ParseUint64Value(ccApp.Entity.Memory)
	application.Name = ccApp.Entity.Name
	application.PackageState = constant.ApplicationPackageState(ccApp.Entity.PackageState)
	application.SpaceGUID = ccApp.Entity.SpaceGUID
	application.StackGUID = ccApp.Entity.StackGUID
	application.StagingFailedDescription = ccApp.Entity.StagingFailedDescription
	application.StagingFailedReason = ccApp.Entity.StagingFailedReason
	application.State = constant.ApplicationState(ccApp.Entity.State)

	if len(ccApp.Entity.EnvironmentVariables) > 0 {
		envVariableValues := map[string]string{}
		for key, value := range ccApp.Entity.EnvironmentVariables {
			envVariableValues[key] = fmt.Sprint(value)
		}
		application.EnvironmentVariables = envVariableValues
	}

	err = application.Instances.ParseStringValue(ccApp.Entity.Instances.String())
	if err != nil {
		return err
	}

	if ccApp.Entity.PackageUpdatedAt != nil {
		application.PackageUpdatedAt = *ccApp.Entity.PackageUpdatedAt
	}
	return nil
}

// CreateApplication creates a cloud controller application in with the given
// settings. SpaceGUID and Name are the only required fields.
func (client *Client) CreateApplication(app Application) (Application, Warnings, error) {
	body, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostAppRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Application{}, nil, err
	}

	var updatedApp Application
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedApp,
	}

	err = client.connection.Make(request, &response)
	return updatedApp, response.Warnings, err
}

// GetApplication returns back an Application.
func (client *Client) GetApplication(guid string) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppRequest,
		URIParams:   Params{"app_guid": guid},
	})
	if err != nil {
		return Application{}, nil, err
	}

	var app Application
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &app,
	}

	err = client.connection.Make(request, &response)
	return app, response.Warnings, err
}

// GetApplications returns back a list of Applications based off of the
// provided filters.
func (client *Client) GetApplications(filters ...Filter) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppsRequest,
		Query:       ConvertFilterParameters(filters),
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

// GetRouteApplications returns a list of Applications based off a route
// GUID and the provided filters.
func (client *Client) GetRouteApplications(routeGUID string, filters ...Filter) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteAppsRequest,
		URIParams:   map[string]string{"route_guid": routeGUID},
		Query:       ConvertFilterParameters(filters),
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

// RestageApplication restages the application with the given GUID.
func (client *Client) RestageApplication(app Application) (Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostAppRestageRequest,
		URIParams:   Params{"app_guid": app.GUID},
	})
	if err != nil {
		return Application{}, nil, err
	}

	var restagedApp Application
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &restagedApp,
	}

	err = client.connection.Make(request, &response)
	return restagedApp, response.Warnings, err
}

// UpdateApplication updates the application with the given GUID. Note: Sending
// DockerImage and StackGUID at the same time will result in an API error.
func (client *Client) UpdateApplication(app Application) (Application, Warnings, error) {
	body, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutAppRequest,
		URIParams:   Params{"app_guid": app.GUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Application{}, nil, err
	}

	var updatedApp Application
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &updatedApp,
	}

	err = client.connection.Make(request, &response)
	return updatedApp, response.Warnings, err
}
