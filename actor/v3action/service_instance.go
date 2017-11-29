package v3action

import (
	"net/url"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type ServiceInstance ccv3.ServiceInstance

func (actor Actor) ShareServiceInstanceInSpaceByOrganizationNameAndSpaceName(serviceInstanceName string, sourceSpaceGUID string, targetOrgName string, targetSpaceName string) (Warnings, error) {
	organization, allWarnings, err := actor.GetOrganizationByName(targetOrgName)
	if err != nil {
		return allWarnings, err
	}

	warnings, err := actor.ShareServiceInstanceInSpaceByOrganizationAndSpaceName(serviceInstanceName, sourceSpaceGUID, organization.GUID, targetSpaceName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}

func (actor Actor) ShareServiceInstanceInSpaceByOrganizationAndSpaceName(serviceInstanceName string, sourceSpaceGUID string, orgGUID string, spaceName string) (Warnings, error) {
	serviceInstance, allWarnings, err := actor.GetServiceInstanceByNameAndSpace(serviceInstanceName, sourceSpaceGUID)

	if _, ok := err.(actionerror.ServiceInstanceNotFoundError); ok == true {
		return allWarnings, actionerror.SharedServiceInstanceNotFound{}
	}

	if err != nil {
		return allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, orgGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	_, apiWarnings, err := actor.CloudControllerClient.ShareServiceInstanceToSpaces(serviceInstance.GUID, []string{space.GUID})
	allWarnings = append(allWarnings, apiWarnings...)

	return allWarnings, err
}

func (actor Actor) GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (ServiceInstance, Warnings, error) {
	serviceInstances, warnings, err := actor.CloudControllerClient.GetServiceInstances(url.Values{
		ccv3.NameFilter:      []string{serviceInstanceName},
		ccv3.SpaceGUIDFilter: []string{spaceGUID},
	})

	if err != nil {
		return ServiceInstance{}, Warnings(warnings), err
	}

	if len(serviceInstances) == 0 {
		return ServiceInstance{}, Warnings(warnings), actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName}
	}

	//Handle multiple serviceInstances being returned as GetServiceInstances arnt filtered by space
	return ServiceInstance(serviceInstances[0]), Warnings(warnings), nil
}
