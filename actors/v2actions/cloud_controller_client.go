package v2actions

import "code.cloudfoundry.org/cli/api/cloudcontrollerv2"

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	GetApplications([]cloudcontrollerv2.Query) ([]cloudcontrollerv2.Application, cloudcontrollerv2.Warnings, error)
	GetServiceInstances([]cloudcontrollerv2.Query) ([]cloudcontrollerv2.ServiceInstance, cloudcontrollerv2.Warnings, error)
	GetServiceBindings([]cloudcontrollerv2.Query) ([]cloudcontrollerv2.ServiceBinding, cloudcontrollerv2.Warnings, error)
	DeleteServiceBinding(serviceBindingGUID string) (cloudcontrollerv2.Warnings, error)
}
