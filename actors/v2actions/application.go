package v2actions

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"
)

type Application struct {
	GUID string
	Name string
}

type ApplicationNotFoundError struct {
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	return fmt.Sprintf("Application '%s' not found.", e.Name)
}

func (actor Actor) GetApplicationBySpace(name string, spaceGUID string) (Application, Warnings, error) {
	app, warnings, err := actor.CloudControllerClient.GetApplications([]cloudcontrollerv2.Query{
		cloudcontrollerv2.Query{
			Filter:   cloudcontrollerv2.NameFilter,
			Operator: cloudcontrollerv2.EqualOperator,
			Value:    name,
		},
		cloudcontrollerv2.Query{
			Filter:   cloudcontrollerv2.SpaceGUIDFilter,
			Operator: cloudcontrollerv2.EqualOperator,
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
