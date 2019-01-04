package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
)

type Service ccv2.Service

func (actor Actor) GetService(serviceGUID string) (Service, Warnings, error) {
	service, warnings, err := actor.CloudControllerClient.GetService(serviceGUID)
	return Service(service), Warnings(warnings), err
}

// GetServiceByName returns a service based on the name provided.
// If there are no services, an ServiceNotFoundError will be returned.
// If there are multiple services, the first will be returned.
func (actor Actor) GetServiceByName(serviceName string) (Service, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetServices(ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	if err != nil {
		return Service{}, Warnings(warnings), err
	}

	if len(services) == 0 {
		return Service{}, Warnings(warnings), actionerror.ServiceNotFoundError{Name: serviceName}
	}

	return Service(services[0]), Warnings(warnings), nil
}

func (actor Actor) getServiceByNameForSpace(serviceName, spaceGUID string) (Service, Warnings, error) {
	services, warnings, err := actor.CloudControllerClient.GetSpaceServices(spaceGUID, ccv2.Filter{
		Type:     constant.LabelFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceName},
	})
	if err != nil {
		return Service{}, Warnings(warnings), err
	}

	if len(services) == 0 {
		return Service{}, Warnings(warnings), actionerror.ServiceNotFoundError{Name: serviceName}
	}

	return Service(services[0]), Warnings(warnings), nil
}
