package v2actions

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

type Application ccv2.Application

type ApplicationNotFoundError struct {
	Name string
}

func (e ApplicationNotFoundError) Error() string {
	return fmt.Sprintf("Application '%s' not found.", e.Name)
}

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
