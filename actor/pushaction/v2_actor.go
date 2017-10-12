package pushaction

import (
	"io"

	"code.cloudfoundry.org/cli/actor/v2action"
)

//go:generate counterfeiter . V2Actor

type V2Actor interface {
	MapRouteToApplication(routeGUID string, appGUID string) (v2action.Warnings, error)
	BindServiceByApplicationAndServiceInstance(appGUID string, serviceInstanceGUID string) (v2action.Warnings, error)
	CreateApplication(application v2action.Application) (v2action.Application, v2action.Warnings, error)
	CreateRoute(route v2action.Route, generatePort bool) (v2action.Route, v2action.Warnings, error)
	FindRouteBoundToSpaceWithSettings(route v2action.Route) (v2action.Route, v2action.Warnings, error)
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
	GetApplicationRoutes(applicationGUID string) (v2action.Routes, v2action.Warnings, error)
	GetDomainsByNameAndOrganization(domainNames []string, orgGUID string) ([]v2action.Domain, v2action.Warnings, error)
	GetOrganizationDomains(orgGUID string) ([]v2action.Domain, v2action.Warnings, error)
	GetServiceInstanceByNameAndSpace(name string, spaceGUID string) (v2action.ServiceInstance, v2action.Warnings, error)
	GetServiceInstancesByApplication(appGUID string) ([]v2action.ServiceInstance, v2action.Warnings, error)
	GetStack(guid string) (v2action.Stack, v2action.Warnings, error)
	GetStackByName(stackName string) (v2action.Stack, v2action.Warnings, error)
	PollJob(job v2action.Job) (v2action.Warnings, error)
	ResourceMatch(allResources []v2action.Resource) ([]v2action.Resource, []v2action.Resource, v2action.Warnings, error)
	UnmapRouteFromApplication(routeGUID string, appGUID string) (v2action.Warnings, error)
	UpdateApplication(application v2action.Application) (v2action.Application, v2action.Warnings, error)
	UploadApplicationPackage(appGUID string, existingResources []v2action.Resource, newResources io.Reader, newResourcesLength int64) (v2action.Job, v2action.Warnings, error)
}
