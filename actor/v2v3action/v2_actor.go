package v2v3action

import "code.cloudfoundry.org/cli/actor/v2action"

//go:generate counterfeiter . V2Actor

type V2Actor interface {
	GetServiceInstanceByNameAndSpace(serviceInstanceName string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error)
	GetServiceInstanceSharedTosByServiceInstance(serviceInstanceGUID string) ([]v2action.ServiceInstanceSharedTo, v2action.Warnings, error)
	GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (v2action.Space, v2action.Warnings, error)
}
