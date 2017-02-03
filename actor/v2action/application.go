package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// Application represents an application.
type Application ccv2.Application

// CalculatedBuildpack returns the buildpack that will be used.
func (application Application) CalculatedBuildpack() string {
	if application.Buildpack != "" {
		return application.Buildpack
	}

	return application.DetectedBuildpack
}

// CalculatedHealthCheckEndpoint returns the health check endpoint.
// If the health check type is not http it will return the empty string.
func (application Application) CalculatedHealthCheckEndpoint() string {
	if application.HealthCheckType == "http" {
		return application.HealthCheckHTTPEndpoint
	}

	return ""
}

// Started returns true when the application is started.
func (application Application) Started() bool {
	return application.State == ccv2.ApplicationStarted
}

// ApplicationNotFoundError is returned when a requested application is not
// found.
type ApplicationNotFoundError struct {
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	return fmt.Sprintf("Application '%s' not found.", e.Name)
}

// GetApplicationByNameAndSpace returns an application with matching name in
// the space.
func (actor Actor) GetApplicationByNameAndSpace(name string, spaceGUID string) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.GetApplications([]ccv2.Query{
		ccv2.Query{
			Filter:   ccv2.NameFilter,
			Operator: ccv2.EqualOperator,
			Value:    name,
		},
		ccv2.Query{
			Filter:   ccv2.SpaceGUIDFilter,
			Operator: ccv2.EqualOperator,
			Value:    spaceGUID,
		},
	})

	if err != nil {
		return Application{}, Warnings(warnings), err
	}

	if len(app) == 0 {
		return Application{}, Warnings(warnings), ApplicationNotFoundError{
			Name: name,
		}
	}

	return Application(app[0]), Warnings(warnings), nil
}

// GetRouteApplications returns a list of apps associated with the provided
// Route GUID.
func (actor Actor) GetRouteApplications(routeGUID string, query []ccv2.Query) ([]Application, Warnings, error) {
	apps, warnings, err := actor.CloudControllerClient.GetRouteApplications(routeGUID, query)
	if err != nil {
		return nil, Warnings(warnings), err
	}
	allApplications := []Application{}
	for _, app := range apps {
		allApplications = append(allApplications, Application(app))
	}
	return allApplications, Warnings(warnings), nil
}

// SetApplicationHealthCheckTypeByNameAndSpace updates an application's health
// check type if it is not already the desired type.
func (actor Actor) SetApplicationHealthCheckTypeByNameAndSpace(name string, spaceGUID string, healthCheckType string, httpEndpoint string) (Application, Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationByNameAndSpace(name, spaceGUID)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return Application{}, allWarnings, err
	}

	if app.HealthCheckType != healthCheckType ||
		healthCheckType == "http" && app.HealthCheckHTTPEndpoint != httpEndpoint {
		var healthCheckHttpEndpoint string
		if healthCheckType == "http" {
			healthCheckHttpEndpoint = httpEndpoint
		}

		updatedApp, apiWarnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
			GUID:                    app.GUID,
			HealthCheckType:         healthCheckType,
			HealthCheckHTTPEndpoint: healthCheckHttpEndpoint,
		})

		allWarnings = append(allWarnings, Warnings(apiWarnings)...)
		return Application(updatedApp), allWarnings, err
	}

	return app, allWarnings, nil
}
