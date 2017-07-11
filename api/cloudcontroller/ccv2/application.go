package ccv2

import (
	"bytes"
	"encoding/json"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

// ApplicationState is the running state of an application.
type ApplicationState string

const (
	ApplicationStarted ApplicationState = "STARTED"
	ApplicationStopped ApplicationState = "STOPPED"
)

// ApplicationPackageState is the staging state of application bits.
type ApplicationPackageState string

const (
	ApplicationPackageStaged  ApplicationPackageState = "STAGED"
	ApplicationPackagePending ApplicationPackageState = "PENDING"
	ApplicationPackageFailed  ApplicationPackageState = "FAILED"
	ApplicationPackageUnknown ApplicationPackageState = "UNKNOWN"
)

// Application represents a Cloud Controller Application.
type Application struct {
	// Buildpack is the buildpack set by the user.
	Buildpack string `json:"buildpack,omitempty"`

	// DetectedBuildpack is the buildpack automatically detected.
	DetectedBuildpack string `json:"-"`

	// DetectedStartCommand is the command used to start the application.
	DetectedStartCommand string `json:"-"`

	// DiskQuota is the disk given to each instance, in megabytes.
	DiskQuota int `json:"-"`

	// DockerImage is the docker image location.
	DockerImage string `json:"docker_image,omitempty"`

	// GUID is the unique application identifier.
	GUID string `json:"guid,omitempty"`

	// HealthCheckType is the type of health check that will be done to the app.
	HealthCheckType string `json:"health_check_type,omitempty"`

	// HealthCheckHTTPEndpoint is the url of the http health check endpoint.
	HealthCheckHTTPEndpoint string `json:"health_check_http_endpoint,omitempty"`

	// Instances is the total number of app instances.
	Instances int `json:"-"`

	// Memory is the memory given to each instance, in megabytes.
	Memory int `json:"-"`

	// Name is the name given to the application.
	Name string `json:"name,omitempty"`

	// PackageState represents the staging state of the application bits.
	PackageState ApplicationPackageState `json:"-"`

	// PackageUpdatedAt is the last time the app bits were updated. In RFC3339.
	PackageUpdatedAt time.Time `json:"-"`

	// SpaceGUID is the GUID of the app's space.
	SpaceGUID string `json:"space_guid,omitempty"`

	// StackGUID is the GUID for the Stack the application is running on.
	StackGUID string `json:"-"`

	// StagingFailedDescription is the verbose description of why the package
	// failed to stage.
	StagingFailedDescription string `json:"-"`

	// StagingFailedReason is the reason why the package failed to stage.
	StagingFailedReason string `json:"-"`

	// State is the desired state of the application.
	State ApplicationState `json:"state,omitempty"`
}

// UnmarshalJSON helps unmarshal a Cloud Controller Application response.
func (application *Application) UnmarshalJSON(data []byte) error {
	var ccApp struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Buildpack                string     `json:"buildpack"`
			DetectedBuildpack        string     `json:"detected_buildpack"`
			DetectedStartCommand     string     `json:"detected_start_command"`
			DiskQuota                int        `json:"disk_quota"`
			DockerImage              string     `json:"docker_image"`
			HealthCheckType          string     `json:"health_check_type"`
			HealthCheckHTTPEndpoint  string     `json:"health_check_http_endpoint"`
			Instances                int        `json:"instances"`
			Memory                   int        `json:"memory"`
			Name                     string     `json:"name"`
			PackageState             string     `json:"package_state"`
			PackageUpdatedAt         *time.Time `json:"package_updated_at"`
			StackGUID                string     `json:"stack_guid"`
			StagingFailedDescription string     `json:"staging_failed_description"`
			StagingFailedReason      string     `json:"staging_failed_reason"`
			State                    string     `json:"state"`
		} `json:"entity"`
	}
	if err := json.Unmarshal(data, &ccApp); err != nil {
		return err
	}

	application.GUID = ccApp.Metadata.GUID
	application.Buildpack = ccApp.Entity.Buildpack
	application.DetectedBuildpack = ccApp.Entity.DetectedBuildpack
	application.DetectedStartCommand = ccApp.Entity.DetectedStartCommand
	application.DiskQuota = ccApp.Entity.DiskQuota
	application.DockerImage = ccApp.Entity.DockerImage
	application.HealthCheckType = ccApp.Entity.HealthCheckType
	application.HealthCheckHTTPEndpoint = ccApp.Entity.HealthCheckHTTPEndpoint
	application.Instances = ccApp.Entity.Instances
	application.Memory = ccApp.Entity.Memory
	application.Name = ccApp.Entity.Name
	application.PackageState = ApplicationPackageState(ccApp.Entity.PackageState)
	application.StackGUID = ccApp.Entity.StackGUID
	application.StagingFailedDescription = ccApp.Entity.StagingFailedDescription
	application.StagingFailedReason = ccApp.Entity.StagingFailedReason
	application.State = ApplicationState(ccApp.Entity.State)

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
		Result: &updatedApp,
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
		Result: &app,
	}

	err = client.connection.Make(request, &response)
	return app, response.Warnings, err
}

// GetApplications returns back a list of Applications based off of the
// provided queries.
func (client *Client) GetApplications(queries []Query) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppsRequest,
		Query:       FormatQueryParameters(queries),
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

// UpdateApplication updates the application with the given GUID.
func (client *Client) UpdateApplication(app Application) (Application, Warnings, error) {
	appGUID := app.GUID
	app.GUID = ""

	body, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PutAppRequest,
		URIParams:   Params{"app_guid": appGUID},
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Application{}, nil, err
	}

	var updatedApp Application
	response := cloudcontroller.Response{
		Result: &updatedApp,
	}

	err = client.connection.Make(request, &response)
	return updatedApp, response.Warnings, err
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
		Result: &restagedApp,
	}

	err = client.connection.Make(request, &response)
	return restagedApp, response.Warnings, err
}

// GetRouteApplications returns a list of Applications associated with a route
// GUID, filtered by provided queries.
func (client *Client) GetRouteApplications(routeGUID string, queryParams []Query) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetRouteAppsRequest,
		URIParams:   map[string]string{"route_guid": routeGUID},
		Query:       FormatQueryParameters(queryParams),
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
