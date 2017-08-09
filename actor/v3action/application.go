package v3action

import (
	"fmt"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/clock"
)

// Application represents a V3 actor application.
type Application ccv3.Application

func (app Application) Started() bool {
	return app.State == "STARTED"
}

// ApplicationNotFoundError represents the error that occurs when the
// application is not found.
type ApplicationNotFoundError struct {
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	return fmt.Sprintf("Application '%s' not found.", e.Name)
}

// ApplicationAlreadyExistsError represents the error that occurs when the
// application already exists.
type ApplicationAlreadyExistsError struct {
	Name string
}

func (e ApplicationAlreadyExistsError) Error() string {
	return fmt.Sprintf("Application '%s' already exists.", e.Name)
}

func (actor Actor) DeleteApplicationByNameAndSpace(name string, spaceGUID string) (Warnings, error) {
	var allWarnings Warnings

	app, getAppWarnings, err := actor.GetApplicationByNameAndSpace(name, spaceGUID)
	allWarnings = append(allWarnings, getAppWarnings...)
	if err != nil {
		return allWarnings, err
	}

	jobURL, deleteAppWarnings, err := actor.CloudControllerClient.DeleteApplication(app.GUID)
	allWarnings = append(allWarnings, deleteAppWarnings...)
	if err != nil {
		return allWarnings, err
	}

	pollWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, pollWarnings...)
	return allWarnings, err
}

// GetApplicationByNameAndSpace returns the application with the given
// name in the given space.
func (actor Actor) GetApplicationByNameAndSpace(appName string, spaceGUID string) (Application, Warnings, error) {
	apps, warnings, err := actor.CloudControllerClient.GetApplications(url.Values{
		"space_guids": []string{spaceGUID},
		"names":       []string{appName},
	})
	if err != nil {
		return Application{}, Warnings(warnings), err
	}

	if len(apps) == 0 {
		return Application{}, Warnings(warnings), ApplicationNotFoundError{Name: appName}
	}

	return Application(apps[0]), Warnings(warnings), nil
}

func (Actor) GetClock() clock.Clock {
	return clock.NewClock()
}

type CreateApplicationInput struct {
	AppName    string
	SpaceGUID  string
	Buildpacks []string
}

// CreateApplicationByNameAndSpace creates and returns the application with the given
// name in the given space.
func (actor Actor) CreateApplicationByNameAndSpace(input CreateApplicationInput) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.CreateApplication(
		ccv3.Application{
			Name: input.AppName,
			Relationships: ccv3.Relationships{
				ccv3.SpaceRelationship: ccv3.Relationship{GUID: input.SpaceGUID},
			},
			Buildpacks: input.Buildpacks,
		})

	if _, ok := err.(ccerror.NameNotUniqueInSpaceError); ok {
		return Application{}, Warnings(warnings), ApplicationAlreadyExistsError{Name: input.AppName}
	}

	return Application(app), Warnings(warnings), err
}

// StopApplication stops an application.
func (actor Actor) StopApplication(appGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.StopApplication(appGUID)

	return Warnings(warnings), err
}

// StartApplication starts an application.
func (actor Actor) StartApplication(appGUID string) (Application, Warnings, error) {
	updatedApp, warnings, err := actor.CloudControllerClient.StartApplication(appGUID)

	return Application(updatedApp), Warnings(warnings), err
}

func (actor Actor) PollStart(appGUID string, warningsChannel chan<- Warnings) error {
	processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(appGUID)
	warningsChannel <- Warnings(warnings)
	if err != nil {
		return err
	}

	timeout := time.Now().Add(actor.Config.StartupTimeout())
	for time.Now().Before(timeout) {
		readyProcs := 0
		for _, process := range processes {
			ready, err := actor.processReady(process, warningsChannel)
			if err != nil {
				return err
			}

			if ready {
				readyProcs++
			}
		}

		if readyProcs == len(processes) {
			return nil
		}
		time.Sleep(actor.Config.PollingInterval())
	}

	return StartupTimeoutError{}
}

// UpdateApplication updates the buildpacks on an application
func (actor Actor) UpdateApplication(appGUID string, buildpacks []string) (Application, Warnings, error) {
	app := ccv3.Application{
		GUID:       appGUID,
		Buildpacks: buildpacks,
	}

	app, warnings, err := actor.CloudControllerClient.UpdateApplication(app)
	return Application(app), Warnings(warnings), err
}

// StartupTimeoutError is returned when startup timeout is reached waiting for
// an application to start.
type StartupTimeoutError struct {
}

func (e StartupTimeoutError) Error() string {
	return fmt.Sprintf("Timed out waiting for application to start")
}

func (actor Actor) processReady(process ccv3.Process, warningsChannel chan<- Warnings) (bool, error) {
	instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(process.GUID)
	warningsChannel <- Warnings(warnings)
	if err != nil {
		return false, err
	}
	if len(instances) == 0 {
		return true, nil
	}

	for _, instance := range instances {
		if instance.State == "RUNNING" {
			return true, nil
		}
	}

	return false, nil
}
