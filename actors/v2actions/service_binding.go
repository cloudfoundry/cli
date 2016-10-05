package v2actions

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontrollerv2"
)

type ServiceBinding struct {
	GUID string
}

type ServiceBindingNotFoundError struct {
	AppGUID             string
	ServiceInstanceGUID string
}

func (e ServiceBindingNotFoundError) Error() string {
	return fmt.Sprintf("Service binding for application GUID '%s', and service instance GUID '%s' not found.", e.AppGUID, e.ServiceInstanceGUID)
}

func (actor Actor) GetServiceBindingByApplicationAndServiceInstance(appGUID string, serviceInstanceGUID string) (ServiceBinding, Warnings, error) {
	serviceBindings, warnings, err := actor.CloudControllerClient.GetServiceBindings([]cloudcontrollerv2.Query{
		cloudcontrollerv2.Query{
			Filter:   cloudcontrollerv2.AppGUIDFilter,
			Operator: cloudcontrollerv2.EqualOperator,
			Value:    appGUID,
		},
		cloudcontrollerv2.Query{
			Filter:   cloudcontrollerv2.ServiceInstanceGUIDFilter,
			Operator: cloudcontrollerv2.EqualOperator,
			Value:    serviceInstanceGUID,
		},
	})

	if err != nil {
		return ServiceBinding{}, Warnings(warnings), err
	}

	if len(serviceBindings) == 0 {
		return ServiceBinding{}, Warnings(warnings), ServiceBindingNotFoundError{
			AppGUID:             appGUID,
			ServiceInstanceGUID: serviceInstanceGUID,
		}
	}

	return ServiceBinding(serviceBindings[0]), Warnings(warnings), err
}

func (actor Actor) UnbindServiceBySpace(appName string, serviceInstanceName string, spaceGUID string) (Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationBySpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	serviceInstance, warnings, err := actor.GetServiceInstanceBySpace(serviceInstanceName, spaceGUID)
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
