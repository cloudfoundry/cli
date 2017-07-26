package v2action

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type ApplicationState string

const (
	ApplicationStateStopping = ApplicationState("stopping")
	ApplicationStateStaging  = ApplicationState("staging")
	ApplicationStateStarting = ApplicationState("starting")
)

// ApplicationInstanceCrashedError is returned when an instance crashes.
type ApplicationInstanceCrashedError struct {
	Name string
}

func (e ApplicationInstanceCrashedError) Error() string {
	return fmt.Sprintf("Application '%s' crashed", e.Name)
}

// ApplicationInstanceFlappingError is returned when an instance crashes.
type ApplicationInstanceFlappingError struct {
	Name string
}

func (e ApplicationInstanceFlappingError) Error() string {
	return fmt.Sprintf("Application '%s' crashed", e.Name)
}

// ApplicationNotFoundError is returned when a requested application is not
// found.
type ApplicationNotFoundError struct {
	GUID string
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	if e.GUID != "" {
		return fmt.Sprintf("Application with GUID '%s' not found.", e.GUID)
	}

	return fmt.Sprintf("Application '%s' not found.", e.Name)
}

// HTTPHealthCheckInvalidError is returned when an HTTP endpoint is used with a
// health check type that is not HTTP.
type HTTPHealthCheckInvalidError struct {
}

func (e HTTPHealthCheckInvalidError) Error() string {
	return "Health check type must be 'http' to set a health check HTTP endpoint"
}

// StagingFailedError is returned when staging an application fails.
type StagingFailedError struct {
	Reason string
}

func (e StagingFailedError) Error() string {
	return e.Reason
}

// StagingFailedNoAppDetectedError is returned when staging an application fails.
type StagingFailedNoAppDetectedError struct {
	Reason string
}

func (e StagingFailedNoAppDetectedError) Error() string {
	return e.Reason
}

// StagingTimeoutError is returned when staging timeout is reached waiting for
// an application to stage.
type StagingTimeoutError struct {
	Name    string
	Timeout time.Duration
}

func (e StagingTimeoutError) Error() string {
	return fmt.Sprintf("Timed out waiting for application '%s' to stage", e.Name)
}

// StartupTimeoutError is returned when startup timeout is reached waiting for
// an application to start.
type StartupTimeoutError struct {
	Name string
}

func (e StartupTimeoutError) Error() string {
	return fmt.Sprintf("Timed out waiting for application '%s' to start", e.Name)
}

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

// StagingCompleted returns true if the application has been staged.
func (application Application) StagingCompleted() bool {
	return application.PackageState == ccv2.ApplicationPackageStaged
}

// StagingFailed returns true if staging the application failed.
func (application Application) StagingFailed() bool {
	return application.PackageState == ccv2.ApplicationPackageFailed
}

// StagingFailedMessage returns the verbose description of the failure or
// the reason if the verbose description is empty.
func (application Application) StagingFailedMessage() string {
	if application.StagingFailedDescription != "" {
		return application.StagingFailedDescription
	}

	return application.StagingFailedReason
}

// StagingFailedNoAppDetected returns true when the staging failed due to a
// NoAppDetectedError.
func (application Application) StagingFailedNoAppDetected() bool {
	return application.StagingFailedReason == "NoAppDetectedError"
}

// Started returns true when the application is started.
func (application Application) Started() bool {
	return application.State == ccv2.ApplicationStarted
}

func (app Application) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpack: '%s', Command: '%s', Detected Buildpack: '%s', Detected Start Command: '%s', Disk Quota: '%d', Docker Image: '%s', Health Check HTTP Endpoint: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Instances: '%d', Memory: '%d', Space GUID: '%s',Stack GUID: '%s', State: '%s'",
		app.Name,
		app.Buildpack,
		app.Command,
		app.DetectedBuildpack,
		app.DetectedStartCommand,
		app.DiskQuota,
		app.DockerImage,
		app.HealthCheckHTTPEndpoint,
		app.HealthCheckTimeout,
		app.HealthCheckType,
		app.Instances,
		app.Memory,
		app.SpaceGUID,
		app.StackGUID,
		app.State,
	)
}

// CreateApplication creates an application.
func (actor Actor) CreateApplication(application Application) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.CreateApplication(ccv2.Application(application))
	return Application(app), Warnings(warnings), err
}

// GetApplication returns the application.
func (actor Actor) GetApplication(guid string) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.GetApplication(guid)

	if _, ok := err.(ccerror.ResourceNotFoundError); ok {
		return Application{}, Warnings(warnings), ApplicationNotFoundError{GUID: guid}
	}

	return Application(app), Warnings(warnings), err
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

// GetApplicationsBySpace returns all applications in a space.
func (actor Actor) GetApplicationsBySpace(spaceGUID string) ([]Application, Warnings, error) {
	ccv2Apps, warnings, err := actor.CloudControllerClient.GetApplications([]ccv2.Query{
		ccv2.Query{
			Filter:   ccv2.SpaceGUIDFilter,
			Operator: ccv2.EqualOperator,
			Value:    spaceGUID,
		},
	})

	if err != nil {
		return []Application{}, Warnings(warnings), err
	}

	apps := make([]Application, len(ccv2Apps))
	for i, ccv2App := range ccv2Apps {
		apps[i] = Application(ccv2App)
	}

	return apps, Warnings(warnings), nil
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
	if httpEndpoint != "/" && healthCheckType != "http" {
		return Application{}, nil, HTTPHealthCheckInvalidError{}
	}

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

// StartApplication restarts a given application. If already stopped, no stop
// call will be sent.
func (actor Actor) StartApplication(app Application, client NOAAClient, config Config) (<-chan *LogMessage, <-chan error, <-chan ApplicationState, <-chan string, <-chan error) {
	messages, logErrs := actor.GetStreamingLogs(app.GUID, client, config)

	appState := make(chan ApplicationState)
	allWarnings := make(chan string)
	errs := make(chan error)
	go func() {
		defer close(appState)
		defer close(allWarnings)
		defer close(errs)
		defer client.Close() // automatic close to prevent stale clients

		if app.PackageState != ccv2.ApplicationPackageStaged {
			appState <- ApplicationStateStaging
		}

		updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
			GUID:  app.GUID,
			State: ccv2.ApplicationStarted,
		})

		for _, warning := range warnings {
			allWarnings <- warning
		}
		if err != nil {
			errs <- err
			return
		}

		actor.waitForApplicationStageAndStart(Application(updatedApp), client, config, appState, allWarnings, errs)
	}()

	return messages, logErrs, appState, allWarnings, errs
}

// RestartApplication restarts a given application. If already stopped, no stop
// call will be sent.
func (actor Actor) RestartApplication(app Application, client NOAAClient, config Config) (<-chan *LogMessage, <-chan error, <-chan ApplicationState, <-chan string, <-chan error) {
	messages, logErrs := actor.GetStreamingLogs(app.GUID, client, config)

	appState := make(chan ApplicationState)
	allWarnings := make(chan string)
	errs := make(chan error)
	go func() {
		defer close(appState)
		defer close(allWarnings)
		defer close(errs)
		defer client.Close() // automatic close to prevent stale clients

		if app.Started() {
			appState <- ApplicationStateStopping
			updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
				GUID:  app.GUID,
				State: ccv2.ApplicationStopped,
			})
			for _, warning := range warnings {
				allWarnings <- warning
			}
			if err != nil {
				errs <- err
				return
			}
			app = Application(updatedApp)
		}

		if app.PackageState != ccv2.ApplicationPackageStaged {
			appState <- ApplicationStateStaging
		}
		updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
			GUID:  app.GUID,
			State: ccv2.ApplicationStarted,
		})

		for _, warning := range warnings {
			allWarnings <- warning
		}
		if err != nil {
			errs <- err
			return
		}

		actor.waitForApplicationStageAndStart(Application(updatedApp), client, config, appState, allWarnings, errs)
	}()

	return messages, logErrs, appState, allWarnings, errs
}

// RestageApplication restarts a given application. If already stopped, no stop
// call will be sent.
func (actor Actor) RestageApplication(app Application, client NOAAClient, config Config) (<-chan *LogMessage, <-chan error, <-chan ApplicationState, <-chan string, <-chan error) {
	messages, logErrs := actor.GetStreamingLogs(app.GUID, client, config)

	appState := make(chan ApplicationState)
	allWarnings := make(chan string)
	errs := make(chan error)
	go func() {
		defer close(appState)
		defer close(allWarnings)
		defer close(errs)
		defer client.Close() // automatic close to prevent stale clients

		appState <- ApplicationStateStaging
		restagedApp, warnings, err := actor.CloudControllerClient.RestageApplication(ccv2.Application{
			GUID: app.GUID,
		})

		for _, warning := range warnings {
			allWarnings <- warning
		}
		if err != nil {
			errs <- err
			return
		}

		actor.waitForApplicationStageAndStart(Application(restagedApp), client, config, appState, allWarnings, errs)
	}()

	return messages, logErrs, appState, allWarnings, errs
}

// UpdateApplication updates an application.
func (actor Actor) UpdateApplication(application Application) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application(application))
	return Application(app), Warnings(warnings), err
}

func (actor Actor) pollStaging(app Application, config Config, allWarnings chan<- string) error {
	timeout := time.Now().Add(config.StagingTimeout())
	for time.Now().Before(timeout) {
		currentApplication, warnings, err := actor.GetApplication(app.GUID)
		for _, warning := range warnings {
			allWarnings <- warning
		}

		switch {
		case err != nil:
			return err
		case currentApplication.StagingCompleted():
			return nil
		case currentApplication.StagingFailed():
			if currentApplication.StagingFailedNoAppDetected() {
				return StagingFailedNoAppDetectedError{Reason: currentApplication.StagingFailedMessage()}
			}
			return StagingFailedError{Reason: currentApplication.StagingFailedMessage()}
		}
		time.Sleep(config.PollingInterval())
	}
	return StagingTimeoutError{Name: app.Name, Timeout: config.StagingTimeout()}
}

func (actor Actor) pollStartup(app Application, config Config, allWarnings chan<- string) error {
	timeout := time.Now().Add(config.StartupTimeout())
	for time.Now().Before(timeout) {
		currentInstances, warnings, err := actor.GetApplicationInstancesByApplication(app.GUID)
		for _, warning := range warnings {
			allWarnings <- warning
		}
		if err != nil {
			return err
		}

		for _, instance := range currentInstances {
			switch {
			case instance.Running():
				return nil
			case instance.Crashed():
				return ApplicationInstanceCrashedError{Name: app.Name}
			case instance.Flapping():
				return ApplicationInstanceFlappingError{Name: app.Name}
			}
		}
		time.Sleep(config.PollingInterval())
	}

	return StartupTimeoutError{Name: app.Name}
}

func (actor Actor) waitForApplicationStageAndStart(app Application, client NOAAClient, config Config, appState chan ApplicationState, allWarnings chan string, errs chan error) {
	err := actor.pollStaging(app, config, allWarnings)
	if err != nil {
		errs <- err
		return
	}

	if app.Instances == 0 {
		return
	}

	client.Close() // Explicit close to stop logs from displaying on the screen
	appState <- ApplicationStateStarting

	err = actor.pollStartup(app, config, allWarnings)
	if err != nil {
		errs <- err
	}
}
