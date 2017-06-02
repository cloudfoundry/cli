package v3action

import (
	"fmt"
	"net/url"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Application represents a V3 actor application.
type Application ccv3.Application

// Process represents a V3 actor process.
type Process struct {
	Type       string
	Instances  []Instance
	MemoryInMB int
}

// Instance represents a V3 actor instance.
type Instance ccv3.Instance

// StartTime returns the time that the instance started.
func (instance *Instance) StartTime() time.Time {
	uptimeDuration := time.Duration(instance.Uptime) * time.Second

	return time.Now().Add(-uptimeDuration)
}

func (p Process) TotalInstanceCount() int {
	return len(p.Instances)
}

func (p Process) HealthyInstanceCount() int {
	count := 0
	for _, instance := range p.Instances {
		if instance.State == "RUNNING" {
			count++
		}
	}
	return count
}

type Buildpack ccv3.Buildpack

type ApplicationSummary struct {
	Application
	Processes      []Process
	CurrentDroplet Droplet
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

// GetApplicationSummaryByNameAndSpace returns an application with process and
// instance stats.
func (actor Actor) GetApplicationSummaryByNameAndSpace(appName string,
	spaceGUID string) (ApplicationSummary, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	ccv3Droplet, warnings, err := actor.CloudControllerClient.GetApplicationCurrentDroplet(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}
	droplet := Droplet{
		Stack: ccv3Droplet.Stack,
	}
	for _, ccv3Buildpack := range ccv3Droplet.Buildpacks {
		droplet.Buildpacks = append(droplet.Buildpacks, Buildpack(ccv3Buildpack))
	}

	ccv3Processes, warnings, err := actor.CloudControllerClient.GetApplicationProcesses(app.GUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return ApplicationSummary{}, allWarnings, err
	}

	var processes []Process
	for _, ccv3Process := range ccv3Processes {
		processGUID := ccv3Process.GUID
		instances, warnings, err := actor.CloudControllerClient.GetProcessInstances(processGUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return ApplicationSummary{}, allWarnings, err
		}

		process := Process{
			Type:       ccv3Process.Type,
			Instances:  []Instance{},
			MemoryInMB: ccv3Process.MemoryInMB,
		}
		for _, instance := range instances {
			process.Instances = append(process.Instances, Instance(instance))
		}

		processes = append(processes, process)
	}

	summary := ApplicationSummary{
		Application:    app,
		Processes:      processes,
		CurrentDroplet: droplet,
	}
	return summary, allWarnings, nil
}

// CreateApplicationByNameAndSpace creates and returns the application with the given
// name in the given space.
func (actor Actor) CreateApplicationByNameAndSpace(appName string, spaceGUID string) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.CreateApplication(
		ccv3.Application{
			Name: appName,
			Relationships: ccv3.Relationships{
				ccv3.SpaceRelationship: ccv3.Relationship{GUID: spaceGUID},
			},
		})

	if _, ok := err.(ccerror.UnprocessableEntityError); ok {
		return Application{}, Warnings(warnings), ApplicationAlreadyExistsError{Name: appName}
	}

	return Application(app), Warnings(warnings), err
}

// SetApplicationDroplet sets the droplet for an application.
func (actor Actor) SetApplicationDroplet(appName string, spaceGUID string, dropletGUID string) (Warnings, error) {
	allWarnings := Warnings{}
	application, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}
	_, apiWarnings, err := actor.CloudControllerClient.SetApplicationDroplet(application.GUID, dropletGUID)
	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	return allWarnings, err
}

// StartApplication starts an application.
func (actor Actor) StartApplication(appName string, spaceGUID string) (Application, Warnings, error) {
	allWarnings := Warnings{}
	application, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return Application{}, allWarnings, err
	}
	updatedApp, apiWarnings, err := actor.CloudControllerClient.StartApplication(application.GUID)
	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	return Application(updatedApp), allWarnings, err
}
