package v2action

import (
	"context"
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type ApplicationStateChange string

const (
	ApplicationStateStopping ApplicationStateChange = "stopping"
	ApplicationStateStaging  ApplicationStateChange = "staging"
	ApplicationStateStarting ApplicationStateChange = "starting"
)

// Application represents an application.
type Application ccv2.Application

// CalculatedBuildpack returns the buildpack that will be used.
func (application Application) CalculatedBuildpack() string {
	if application.Buildpack.IsSet {
		return application.Buildpack.Value
	}

	return application.DetectedBuildpack.Value
}

// CalculatedCommand returns the command that will be used.
func (application Application) CalculatedCommand() string {
	if application.Command.IsSet {
		return application.Command.Value
	}

	return application.DetectedStartCommand.Value
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
	return application.PackageState == constant.ApplicationPackageStaged
}

// StagingFailed returns true if staging the application failed.
func (application Application) StagingFailed() bool {
	return application.PackageState == constant.ApplicationPackageFailed
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
	return application.State == constant.ApplicationStarted
}

// Stopped returns true when the application is stopped.
func (application Application) Stopped() bool {
	return application.State == constant.ApplicationStopped
}

func (application Application) String() string {
	return fmt.Sprintf(
		"App Name: '%s', Buildpack (%t, '%s'), Command: (%t, '%s'), Detected Buildpack: (%t, '%s'), Detected Start Command: (%t, '%s'), Disk Quota: '%d', Docker Image: '%s', Health Check HTTP Endpoint: '%s', Health Check Timeout: '%d', Health Check Type: '%s', Instances: (%t, '%d'), Memory: '%d', Space GUID: '%s', Stack GUID: '%s', State: '%s'",
		application.Name,
		application.Buildpack.IsSet,
		application.Buildpack.Value,
		application.Command.IsSet,
		application.Command.Value,
		application.DetectedBuildpack.IsSet,
		application.DetectedBuildpack.Value,
		application.DetectedStartCommand.IsSet,
		application.DetectedStartCommand.Value,
		application.DiskQuota.Value,
		application.DockerImage,
		application.HealthCheckHTTPEndpoint,
		application.HealthCheckTimeout,
		application.HealthCheckType,
		application.Instances.IsSet,
		application.Instances.Value,
		application.Memory.Value,
		application.SpaceGUID,
		application.StackGUID,
		application.State,
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
		return Application{}, Warnings(warnings), actionerror.ApplicationNotFoundError{GUID: guid}
	}

	return Application(app), Warnings(warnings), err
}

// GetApplicationByNameAndSpace returns an application with matching name in
// the space.
func (actor Actor) GetApplicationByNameAndSpace(name string, spaceGUID string) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.GetApplications(
		ccv2.Filter{
			Type:     constant.NameFilter,
			Operator: constant.EqualOperator,
			Values:   []string{name},
		},
		ccv2.Filter{
			Type:     constant.SpaceGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{spaceGUID},
		},
	)

	if err != nil {
		return Application{}, Warnings(warnings), err
	}

	if len(app) == 0 {
		return Application{}, Warnings(warnings), actionerror.ApplicationNotFoundError{
			Name: name,
		}
	}

	return Application(app[0]), Warnings(warnings), nil
}

// GetApplicationsBySpace returns all applications in a space.
func (actor Actor) GetApplicationsBySpace(spaceGUID string) ([]Application, Warnings, error) {
	ccv2Apps, warnings, err := actor.CloudControllerClient.GetApplications(
		ccv2.Filter{
			Type:     constant.SpaceGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{spaceGUID},
		},
	)

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
func (actor Actor) GetRouteApplications(routeGUID string) ([]Application, Warnings, error) {
	apps, warnings, err := actor.CloudControllerClient.GetRouteApplications(routeGUID)
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
func (actor Actor) SetApplicationHealthCheckTypeByNameAndSpace(name string, spaceGUID string, healthCheckType constant.ApplicationHealthCheckType, httpEndpoint string) (Application, Warnings, error) {
	if httpEndpoint != "/" && healthCheckType != constant.ApplicationHealthCheckHTTP {
		return Application{}, nil, actionerror.HTTPHealthCheckInvalidError{}
	}

	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationByNameAndSpace(name, spaceGUID)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return Application{}, allWarnings, err
	}

	if app.HealthCheckType != healthCheckType ||
		healthCheckType == constant.ApplicationHealthCheckHTTP && app.HealthCheckHTTPEndpoint != httpEndpoint {
		var healthCheckEndpoint string
		if healthCheckType == constant.ApplicationHealthCheckHTTP {
			healthCheckEndpoint = httpEndpoint
		}

		updatedApp, apiWarnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
			GUID:                    app.GUID,
			HealthCheckType:         healthCheckType,
			HealthCheckHTTPEndpoint: healthCheckEndpoint,
		})

		allWarnings = append(allWarnings, Warnings(apiWarnings)...)
		return Application(updatedApp), allWarnings, err
	}

	return app, allWarnings, nil
}

// StartApplication restarts a given application. If already stopped, no stop
// call will be sent.

func (actor Actor) StartApplication(app Application, client LogCacheClient) (<-chan *LogMessage, <-chan error, <-chan ApplicationStateChange, <-chan string, <-chan error) {
	messages, logErrs, stopStreaming := actor.GetStreamingLogs(app.GUID, client)

	appState := make(chan ApplicationStateChange)
	allWarnings := make(chan string)
	errs := make(chan error)
	go func() {
		defer close(appState)
		defer close(allWarnings)
		defer close(errs)
		defer stopStreaming()

		if app.PackageState != constant.ApplicationPackageStaged {
			appState <- ApplicationStateStaging
		}

		updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
			GUID:  app.GUID,
			State: constant.ApplicationStarted,
		})

		for _, warning := range warnings {
			allWarnings <- warning
		}
		if err != nil {
			errs <- err
			return
		}

		actor.waitForApplicationStageAndStart(Application(updatedApp), stopStreaming, appState, allWarnings, errs)
	}()

	return messages, logErrs, appState, allWarnings, errs
}

// RestartApplication restarts a given application. If already stopped, no stop
// call will be sent.
func (actor Actor) RestartApplication(app Application, client LogCacheClient) (<-chan *LogMessage, <-chan error, <-chan ApplicationStateChange, <-chan string, <-chan error) {
	messages, logErrs, stopStreaming := actor.GetStreamingLogs(app.GUID, nil)

	appState := make(chan ApplicationStateChange)
	allWarnings := make(chan string)
	errs := make(chan error)
	go func() {
		defer close(appState)
		defer close(allWarnings)
		defer close(errs)
		defer stopStreaming()

		if app.Started() {
			appState <- ApplicationStateStopping
			updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
				GUID:  app.GUID,
				State: constant.ApplicationStopped,
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

		if app.PackageState != constant.ApplicationPackageStaged {
			appState <- ApplicationStateStaging
		}
		updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application{
			GUID:  app.GUID,
			State: constant.ApplicationStarted,
		})

		for _, warning := range warnings {
			allWarnings <- warning
		}
		if err != nil {
			errs <- err
			return
		}

		actor.waitForApplicationStageAndStart(Application(updatedApp), stopStreaming, appState, allWarnings, errs)
	}()

	return messages, logErrs, appState, allWarnings, errs
}

// RestageApplication restarts a given application. If already stopped, no stop
// call will be sent.
func (actor Actor) RestageApplication(app Application, client LogCacheClient) (<-chan *LogMessage, <-chan error, <-chan ApplicationStateChange, <-chan string, <-chan error) {
	messages, logErrs, stopStreaming := actor.GetStreamingLogs(app.GUID, nil)

	appState := make(chan ApplicationStateChange)
	allWarnings := make(chan string)
	errs := make(chan error)
	go func() {
		defer close(appState)
		defer close(allWarnings)
		defer close(errs)
		defer stopStreaming()

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

		actor.waitForApplicationStageAndStart(Application(restagedApp), stopStreaming, appState, allWarnings, errs)
	}()

	return messages, logErrs, appState, allWarnings, errs
}

// UpdateApplication updates an application.
func (actor Actor) UpdateApplication(application Application) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.UpdateApplication(ccv2.Application(application))
	return Application(app), Warnings(warnings), err
}

func (actor Actor) pollStaging(app Application, allWarnings chan<- string) error {
	timeout := time.Now().Add(actor.Config.StagingTimeout())
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
				return actionerror.StagingFailedNoAppDetectedError{Reason: currentApplication.StagingFailedMessage()}
			}
			return actionerror.StagingFailedError{Reason: currentApplication.StagingFailedMessage()}
		}
		time.Sleep(actor.Config.PollingInterval())
	}
	return actionerror.StagingTimeoutError{AppName: app.Name, Timeout: actor.Config.StagingTimeout()}
}

func (actor Actor) pollStartup(app Application, allWarnings chan<- string) error {
	timeout := time.Now().Add(actor.Config.StartupTimeout())
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
				return actionerror.ApplicationInstanceCrashedError{Name: app.Name}
			case instance.Flapping():
				return actionerror.ApplicationInstanceFlappingError{Name: app.Name}
			}
		}
		time.Sleep(actor.Config.PollingInterval())
	}

	return actionerror.StartupTimeoutError{Name: app.Name}
}

func (actor Actor) waitForApplicationStageAndStart(app Application, stopStreaming context.CancelFunc, appState chan ApplicationStateChange, allWarnings chan string, errs chan error) {
	err := actor.pollStaging(app, allWarnings)
	if err != nil {
		errs <- err
		return
	}

	if app.Instances.Value == 0 {
		return
	}

	stopStreaming()
	appState <- ApplicationStateStarting

	err = actor.pollStartup(app, allWarnings)
	if err != nil {
		errs <- err
	}
}
