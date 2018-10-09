package v2action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
)

// ServiceKey represents a set of credentials for a service instance.
type ServiceKey ccv2.ServiceKey

func (actor Actor) CreateServiceKey(serviceInstanceName, keyName, spaceGUID string, parameters map[string]interface{}) (ServiceKey, Warnings, error) {
	var allWarnings Warnings

	serviceInstance, warnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return ServiceKey{}, allWarnings, err
	}

	serviceKey, ccv2Warnings, err := actor.CloudControllerClient.CreateServiceKey(serviceInstance.GUID, keyName, parameters)
	allWarnings = append(allWarnings, ccv2Warnings...)

	return ServiceKey(serviceKey), allWarnings, err
}
