package v3action

import (
	"fmt"
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

// Application represents a V3 actor application.
type Application ccv3.Application

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
