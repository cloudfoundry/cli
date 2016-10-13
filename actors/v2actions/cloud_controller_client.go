package v2actions

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	DeleteServiceBinding(serviceBindingGUID string) (ccv2.Warnings, error)
	GetApplications([]ccv2.Query) ([]ccv2.Application, ccv2.Warnings, error)
	GetServiceBindings([]ccv2.Query) ([]ccv2.ServiceBinding, ccv2.Warnings, error)
	GetServiceInstances([]ccv2.Query) ([]ccv2.ServiceInstance, ccv2.Warnings, error)
	GetSpaceServiceInstances(spaceGUID string, includeUserProvidedServices bool, queries []ccv2.Query) ([]ccv2.ServiceInstance, ccv2.Warnings, error)
}
