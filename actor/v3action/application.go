package v3action

import (
	"fmt"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Application represents a V3 actor application.
type Application struct {
	Name      string
	GUID      string
	State     string
	Lifecycle AppLifecycle
}

type AppLifecycle struct {
	Type AppLifecycleType
	Data AppLifecycleData
}

type AppLifecycleType ccv3.AppLifecycleType
type AppLifecycleData ccv3.AppLifecycleData

const (
	BuildpackAppLifecycleType AppLifecycleType = AppLifecycleType(ccv3.BuildpackAppLifecycleType)
	DockerAppLifecycleType    AppLifecycleType = AppLifecycleType(ccv3.DockerAppLifecycleType)
)

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

	return Application{
		Name:  apps[0].Name,
		GUID:  apps[0].GUID,
		State: apps[0].State,
		Lifecycle: AppLifecycle{
			Type: AppLifecycleType(apps[0].Lifecycle.Type),
			Data: AppLifecycleData(apps[0].Lifecycle.Data),
		},
	}, Warnings(warnings), nil
}

// GetApplicationsBySpace returns all applications in a space.
func (actor Actor) GetApplicationsBySpace(spaceGUID string) ([]Application, Warnings, error) {
	ccv3Apps, warnings, err := actor.CloudControllerClient.GetApplications(url.Values{
		"space_guids": []string{spaceGUID},
	})

	if err != nil {
		return []Application{}, Warnings(warnings), err
	}

	apps := make([]Application, len(ccv3Apps))
	for i, ccv3App := range ccv3Apps {
		apps[i] = Application{
			Name:  ccv3App.Name,
			GUID:  ccv3App.GUID,
			State: ccv3App.State,
			Lifecycle: AppLifecycle{
				Type: AppLifecycleType(ccv3App.Lifecycle.Type),
				Data: AppLifecycleData(ccv3App.Lifecycle.Data),
			},
		}
	}
	return apps, Warnings(warnings), nil
}

// CreateApplicationInSpace creates and returns the application with the given
// name in the given space.
func (actor Actor) CreateApplicationInSpace(app Application, spaceGUID string) (Application, Warnings, error) {
	createdApp, warnings, err := actor.CloudControllerClient.CreateApplication(
		ccv3.Application{
			Name: app.Name,
			Relationships: ccv3.Relationships{
				ccv3.SpaceRelationship: ccv3.Relationship{GUID: spaceGUID},
			},
			Lifecycle: ccv3.AppLifecycle{
				Type: ccv3.AppLifecycleType(app.Lifecycle.Type),
				Data: ccv3.AppLifecycleData{
					Buildpacks: app.Lifecycle.Data.Buildpacks,
				},
			},
		})

	if err != nil {
		if _, ok := err.(ccerror.NameNotUniqueInSpaceError); ok {
			return Application{}, Warnings(warnings), ApplicationAlreadyExistsError{Name: app.Name}
		}
		return Application{}, Warnings(warnings), err
	}

	return Application{
		Name:  createdApp.Name,
		GUID:  createdApp.GUID,
		State: createdApp.State,
		Lifecycle: AppLifecycle{
			Type: AppLifecycleType(createdApp.Lifecycle.Type),
			Data: AppLifecycleData(createdApp.Lifecycle.Data),
		},
	}, Warnings(warnings), nil
}

// StopApplication stops an application.
func (actor Actor) StopApplication(appGUID string) (Warnings, error) {
	warnings, err := actor.CloudControllerClient.StopApplication(appGUID)

	return Warnings(warnings), err
}

// StartApplication starts an application.
func (actor Actor) StartApplication(appGUID string) (Application, Warnings, error) {
	updatedApp, warnings, err := actor.CloudControllerClient.StartApplication(appGUID)
	if err != nil {
		return Application{}, Warnings(warnings), err
	}

	return Application{
		Name:  updatedApp.Name,
		GUID:  updatedApp.GUID,
		State: updatedApp.State,
		Lifecycle: AppLifecycle{
			Type: AppLifecycleType(updatedApp.Lifecycle.Type),
			Data: AppLifecycleData(updatedApp.Lifecycle.Data),
		},
	}, Warnings(warnings), nil
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
			ready, err := actor.processStatus(process, warningsChannel)
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
func (actor Actor) UpdateApplication(app Application) (Application, Warnings, error) {
	ccApp := ccv3.Application{
		GUID: app.GUID,
		Lifecycle: ccv3.AppLifecycle{
			Type: ccv3.AppLifecycleType(app.Lifecycle.Type),
			Data: ccv3.AppLifecycleData(app.Lifecycle.Data),
		},
	}

	updatedApp, warnings, err := actor.CloudControllerClient.UpdateApplication(ccApp)
	if err != nil {
		return Application{}, Warnings(warnings), err
	}

	return Application{
		Name:  updatedApp.Name,
		GUID:  updatedApp.GUID,
		State: updatedApp.State,
		Lifecycle: AppLifecycle{
			Type: AppLifecycleType(updatedApp.Lifecycle.Type),
			Data: AppLifecycleData(updatedApp.Lifecycle.Data),
		},
	}, Warnings(warnings), nil
}

// StartupTimeoutError is returned when startup timeout is reached waiting for
// an application to start.
type StartupTimeoutError struct {
}

func (e StartupTimeoutError) Error() string {
	return fmt.Sprintf("Timed out waiting for application to start")
}

func (actor Actor) processStatus(process ccv3.Process, warningsChannel chan<- Warnings) (bool, error) {
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

	for _, instance := range instances {
		if instance.State != "CRASHED" {
			return false, nil
		}
	}

	// all of the instances are crashed at this point
	return true, nil
}
