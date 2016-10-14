package v2actions

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

//go:generate counterfeiter . CloudControllerClient

type CloudControllerClient interface {
	DeleteRoute(routeGUID string) (ccv2.Warnings, error)
	DeleteServiceBinding(serviceBindingGUID string) (ccv2.Warnings, error)
	GetApplications(queries []ccv2.Query) ([]ccv2.Application, ccv2.Warnings, error)
	GetPrivateDomain(domainGUID string) (ccv2.Domain, ccv2.Warnings, error)
	GetRouteApplications(routeGUID string, queries []ccv2.Query) ([]ccv2.Application, ccv2.Warnings, error)
	GetServiceBindings(queries []ccv2.Query) ([]ccv2.ServiceBinding, ccv2.Warnings, error)
	GetServiceInstances(queries []ccv2.Query) ([]ccv2.ServiceInstance, ccv2.Warnings, error)
	GetSharedDomain(domainGUID string) (ccv2.Domain, ccv2.Warnings, error)
	GetSpaceRoutes(spaceGUID string, queries []ccv2.Query) ([]ccv2.Route, ccv2.Warnings, error)
	GetSpaceServiceInstances(spaceGUID string, includeUserProvidedServices bool, queries []ccv2.Query) ([]ccv2.ServiceInstance, ccv2.Warnings, error)
}
