package ccv2

import (
	"bytes"
	"encoding/json"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/internal"
)

type ApplicationState string

const (
	ApplicationStarted ApplicationState = "STARTED"
	ApplicationStopped                  = "STOPPED"
)

// Application represents a Cloud Controller Application.
type Application struct {
	// Buildpack is the buildpack set by the user.
	Buildpack string `json:"-"`

	// DetectedBuildpack is the buildpack automatically detected.
	DetectedBuildpack string `json:"-"`

	// DetectedStartCommand is the command used to start the application.
	DetectedStartCommand string `json:"-"`

	// DiskQuota is the disk given to each instance, in megabytes.
	DiskQuota int `json:"-"`

	// GUID is the unique application identifier.
	GUID string `json:"-"`

	// HealthCheckType is the type of health check that will be done to the app.
	HealthCheckType string `json:"health_check_type,omitempty"`

	// HealthCheckHTTPEndpoint is the url of the http health check endpoint.
	HealthCheckHTTPEndpoint string `json:"health_check_http_endpoint,omitempty"`

	// Instances is the total number of app instances.
	Instances int `json:"-"`

	// Memory is the memory given to each instance, in megabytes.
	Memory int `json:"-"`

	// Name is the name given to the application.
	Name string `json:"-"`

	// PackageUpdatedAt is the last time the app bits were updated. In RFC3339.
	PackageUpdatedAt time.Time `json:"-"`

	// StackGUID is the GUID for the Stack the application is running on.
	StackGUID string `json:"-"`

	// State is the desired state of the application.
	State ApplicationState `json:"-"`
}

// UnmarshalJSON helps unmarshal a Cloud Controller Application response.
func (application *Application) UnmarshalJSON(data []byte) error {
	var ccApp struct {
		Metadata internal.Metadata `json:"metadata"`
		Entity   struct {
			Buildpack               string     `json:"buildpack"`
			DetectedBuildpack       string     `json:"detected_buildpack"`
			DetectedStartCommand    string     `json:"detected_start_command"`
			DiskQuota               int        `json:"disk_quota"`
			HealthCheckType         string     `json:"health_check_type"`
			HealthCheckHTTPEndpoint string     `json:"health_check_http_endpoint"`
			Instances               int        `json:"instances"`
			Memory                  int        `json:"memory"`
			Name                    string     `json:"name"`
			PackageUpdatedAt        *time.Time `json:"package_updated_at"`
			StackGUID               string     `json:"stack_guid"`
			State                   string     `json:"state"`
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
	application.HealthCheckType = ccApp.Entity.HealthCheckType
	application.HealthCheckHTTPEndpoint = ccApp.Entity.HealthCheckHTTPEndpoint
	application.Instances = ccApp.Entity.Instances
	application.Memory = ccApp.Entity.Memory
	application.Name = ccApp.Entity.Name
	application.StackGUID = ccApp.Entity.StackGUID
	application.State = ApplicationState(ccApp.Entity.State)

	if ccApp.Entity.PackageUpdatedAt != nil {
		application.PackageUpdatedAt = *ccApp.Entity.PackageUpdatedAt
	}
	return nil
}

// GetApplications returns back a list of Applications based off of the
// provided queries.
func (client *Client) GetApplications(queries []Query) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.AppsRequest,
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
			return cloudcontroller.UnknownObjectInListError{
				Expected:   Application{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullAppsList, warnings, err
}

func (client *Client) UpdateApplication(app Application) (Application, Warnings, error) {
	body, err := json.Marshal(app)
	if err != nil {
		return Application{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.UpdateAppRequest,
		URIParams:   Params{"app_guid": app.GUID},
		Body:        bytes.NewBuffer(body),
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

// GetRouteApplications returns a list of Applications associated with a route
// GUID, filtered by provided queries.
func (client *Client) GetRouteApplications(routeGUID string, queryParams []Query) ([]Application, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.AppsFromRouteRequest,
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
			return cloudcontroller.UnknownObjectInListError{
				Expected:   Application{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullAppsList, warnings, err
}
