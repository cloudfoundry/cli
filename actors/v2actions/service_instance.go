package v2actions

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"
)

type ServiceInstance struct {
	GUID string
	Name string
}

type ServiceInstanceNotFoundError struct {
	Name string
}

func (e ServiceInstanceNotFoundError) Error() string {
	return fmt.Sprintf("Service instance '%s' not found.", e.Name)
}

func (actor Actor) GetServiceInstanceBySpace(name string, spaceGUID string) (ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetServiceInstances([]cloudcontrollerv2.Query{
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
		return ServiceInstance{}, Warnings(warnings), err
	}

	if len(serviceInstances) == 0 {
		return ServiceInstance{}, Warnings(warnings), ServiceInstanceNotFoundError{
			Name: name,
		}
	}

	return ServiceInstance(serviceInstances[0]), Warnings(warnings), nil
}
