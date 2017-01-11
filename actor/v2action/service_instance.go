package v2action

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// ServiceInstance represents an instance of a service.
type ServiceInstance ccv2.ServiceInstance

type ServiceInstanceNotFoundError struct {
	Name string
}

func (e ServiceInstanceNotFoundError) Error() string {
	return fmt.Sprintf("Service instance '%s' not found.", e.Name)
}

// GetServiceInstanceByNameAndSpace returns a given instance by name and space
// GUID.
func (actor Actor) GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetServiceInstances([]ccv2.Query{
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
		return ServiceInstance{}, Warnings(warnings), err
	}

	if len(serviceInstances) == 0 {
		return ServiceInstance{}, Warnings(warnings), ServiceInstanceNotFoundError{
			Name: name,
		}
	}

	return ServiceInstance(serviceInstances[0]), Warnings(warnings), nil
}

// TODO: Delete Me and use the other one
func (actor Actor) GetSpaceServiceInstanceByName(spaceGUID string, name string) (ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetSpaceServiceInstances(
		spaceGUID,
		true,
		[]ccv2.Query{
			ccv2.Query{
				Filter:   ccv2.NameFilter,
				Operator: ccv2.EqualOperator,
				Value:    name,
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
