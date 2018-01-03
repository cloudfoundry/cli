package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// ServiceBinding represents the link between a service instance and an
// application.
type ServiceBinding ccv2.ServiceBinding

// BindServiceByApplicationAndServiceInstance binds the service instance to an application.
func (actor Actor) BindServiceByApplicationAndServiceInstance(appGUID string, serviceInstanceGUID string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.CreateServiceBinding(appGUID, serviceInstanceGUID, nil)

	return Warnings(warnings), err
}

// BindServiceBySpace binds the service instance to an application for a given space.
func (actor Actor) BindServiceBySpace(appName string, serviceInstanceName string, spaceGUID string, parameters map[string]interface{}) (Warnings, error) {
	var allWarnings Warnings
	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	_, ccv2Warnings, err := actor.CloudControllerClient.CreateServiceBinding(app.GUID, serviceInstance.GUID, parameters)
	allWarnings = append(allWarnings, ccv2Warnings...)

	return allWarnings, err
}

// GetServiceBindingByApplicationAndServiceInstance returns a service binding
// given an application GUID and and service instance GUID.
func (actor Actor) GetServiceBindingByApplicationAndServiceInstance(appGUID string, serviceInstanceGUID string) (ServiceBinding, Warnings, error) {
	serviceBindings, warnings, err := actor.CloudControllerClient.GetServiceBindings(
		ccv2.QQuery{
			Filter:   ccv2.AppGUIDFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{appGUID},
		},
		ccv2.QQuery{
			Filter:   ccv2.ServiceInstanceGUIDFilter,
			Operator: ccv2.EqualOperator,
			Values:   []string{serviceInstanceGUID},
		},
	)

	if err != nil {
		return ServiceBinding{}, Warnings(warnings), err
	}

	if len(serviceBindings) == 0 {
		return ServiceBinding{}, Warnings(warnings), actionerror.ServiceBindingNotFoundError{
			AppGUID:             appGUID,
			ServiceInstanceGUID: serviceInstanceGUID,
		}
	}

	return ServiceBinding(serviceBindings[0]), Warnings(warnings), err
}

// UnbindServiceBySpace deletes the service binding between an application and
// service instance for a given space.
func (actor Actor) UnbindServiceBySpace(appName string, serviceInstanceName string, spaceGUID string) (Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	serviceBinding, warnings, err := actor.GetServiceBindingByApplicationAndServiceInstance(app.GUID, serviceInstance.GUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	ccWarnings, err := actor.CloudControllerClient.DeleteServiceBinding(serviceBinding.GUID)
	allWarnings = append(allWarnings, ccWarnings...)

	return allWarnings, err
}

func (actor Actor) GetServiceBindingsByServiceInstance(serviceInstanceGUID string) ([]ServiceBinding, Warnings, error) {
	serviceBindings, warnings, err := actor.CloudControllerClient.GetServiceInstanceServiceBindings(serviceInstanceGUID)
	if err != nil {
		return nil, Warnings(warnings), err
	}

	allServiceBindings := []ServiceBinding{}
	for _, serviceBinding := range serviceBindings {
		allServiceBindings = append(allServiceBindings, ServiceBinding(serviceBinding))
	}

	return allServiceBindings, Warnings(warnings), nil
}

func (actor Actor) GetServiceBindingsByUserProvidedServiceInstance(userProvidedServiceInstanceGUID string) ([]ServiceBinding, Warnings, error) {
	serviceBindings, warnings, err := actor.CloudControllerClient.GetUserProvidedServiceInstanceServiceBindings(userProvidedServiceInstanceGUID)
	if err != nil {
		return nil, Warnings(warnings), err
	}
	allServiceBindings := []ServiceBinding{}
	for _, serviceBinding := range serviceBindings {
		allServiceBindings = append(allServiceBindings, ServiceBinding(serviceBinding))
	}

	return allServiceBindings, Warnings(warnings), nil
}
