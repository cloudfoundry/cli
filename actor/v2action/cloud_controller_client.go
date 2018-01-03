package v2action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

//go:generate counterfeiter . CloudControllerClient

// CloudControllerClient is a Cloud Controller V2 client.
type CloudControllerClient interface {
	AssociateSpaceWithRunningSecurityGroup(securityGroupGUID string, spaceGUID string) (ccv2.Warnings, error)
	AssociateSpaceWithStagingSecurityGroup(securityGroupGUID string, spaceGUID string) (ccv2.Warnings, error)
	CheckRoute(route ccv2.Route) (bool, ccv2.Warnings, error)
	CreateApplication(app ccv2.Application) (ccv2.Application, ccv2.Warnings, error)
	CreateRoute(route ccv2.Route, generatePort bool) (ccv2.Route, ccv2.Warnings, error)
	CreateServiceBinding(appGUID string, serviceBindingGUID string, parameters map[string]interface{}) (ccv2.ServiceBinding, ccv2.Warnings, error)
	CreateUser(uaaUserID string) (ccv2.User, ccv2.Warnings, error)
	DeleteOrganization(orgGUID string) (ccv2.Job, ccv2.Warnings, error)
	DeleteRoute(routeGUID string) (ccv2.Warnings, error)
	DeleteRouteApplication(routeGUID string, appGUID string) (ccv2.Warnings, error)
	DeleteServiceBinding(serviceBindingGUID string) (ccv2.Warnings, error)
	DeleteSpace(spaceGUID string) (ccv2.Job, ccv2.Warnings, error)
	GetApplication(guid string) (ccv2.Application, ccv2.Warnings, error)
	GetApplicationInstanceStatusesByApplication(guid string) (map[int]ccv2.ApplicationInstanceStatus, ccv2.Warnings, error)
	GetApplicationInstancesByApplication(guid string) (map[int]ccv2.ApplicationInstance, ccv2.Warnings, error)
	GetApplicationRoutes(appGUID string, queries ...ccv2.QQuery) ([]ccv2.Route, ccv2.Warnings, error)
	GetApplications(queries ...ccv2.QQuery) ([]ccv2.Application, ccv2.Warnings, error)
	GetConfigFeatureFlags() ([]ccv2.FeatureFlag, ccv2.Warnings, error)
	GetJob(jobGUID string) (ccv2.Job, ccv2.Warnings, error)
	GetOrganization(guid string) (ccv2.Organization, ccv2.Warnings, error)
	GetOrganizationPrivateDomains(orgGUID string, queries ...ccv2.QQuery) ([]ccv2.Domain, ccv2.Warnings, error)
	GetOrganizationQuota(guid string) (ccv2.OrganizationQuota, ccv2.Warnings, error)
	GetOrganizations(queries ...ccv2.QQuery) ([]ccv2.Organization, ccv2.Warnings, error)
	GetPrivateDomain(domainGUID string) (ccv2.Domain, ccv2.Warnings, error)
	GetRouteApplications(routeGUID string, queries ...ccv2.QQuery) ([]ccv2.Application, ccv2.Warnings, error)
	GetRoutes(queries ...ccv2.QQuery) ([]ccv2.Route, ccv2.Warnings, error)
	GetRunningSpacesBySecurityGroup(securityGroupGUID string) ([]ccv2.Space, ccv2.Warnings, error)
	GetSecurityGroups(queries ...ccv2.QQuery) ([]ccv2.SecurityGroup, ccv2.Warnings, error)
	GetService(serviceGUID string) (ccv2.Service, ccv2.Warnings, error)
	GetServiceBindings(queries ...ccv2.QQuery) ([]ccv2.ServiceBinding, ccv2.Warnings, error)
	GetServiceInstance(serviceInstanceGUID string) (ccv2.ServiceInstance, ccv2.Warnings, error)
	GetServiceInstanceServiceBindings(serviceInstanceGUID string) ([]ccv2.ServiceBinding, ccv2.Warnings, error)
	GetServiceInstanceSharedFrom(serviceInstanceGUID string) (ccv2.ServiceInstanceSharedFrom, ccv2.Warnings, error)
	GetServiceInstanceSharedTos(serviceInstanceGUID string) ([]ccv2.ServiceInstanceSharedTo, ccv2.Warnings, error)
	GetUserProvidedServiceInstanceServiceBindings(userProvidedServiceInstanceGUID string) ([]ccv2.ServiceBinding, ccv2.Warnings, error)
	GetServiceInstances(queries ...ccv2.QQuery) ([]ccv2.ServiceInstance, ccv2.Warnings, error)
	GetServicePlan(servicePlanGUID string) (ccv2.ServicePlan, ccv2.Warnings, error)
	GetSharedDomain(domainGUID string) (ccv2.Domain, ccv2.Warnings, error)
	GetSharedDomains(queries ...ccv2.QQuery) ([]ccv2.Domain, ccv2.Warnings, error)
	GetSpaceQuota(guid string) (ccv2.SpaceQuota, ccv2.Warnings, error)
	GetSpaceRoutes(spaceGUID string, queries ...ccv2.QQuery) ([]ccv2.Route, ccv2.Warnings, error)
	GetSpaceRunningSecurityGroupsBySpace(spaceGUID string, queries ...ccv2.QQuery) ([]ccv2.SecurityGroup, ccv2.Warnings, error)
	GetSpaceServiceInstances(spaceGUID string, includeUserProvidedServices bool, queries ...ccv2.QQuery) ([]ccv2.ServiceInstance, ccv2.Warnings, error)
	GetSpaceStagingSecurityGroupsBySpace(spaceGUID string, queries ...ccv2.QQuery) ([]ccv2.SecurityGroup, ccv2.Warnings, error)
	GetSpaces(queries ...ccv2.QQuery) ([]ccv2.Space, ccv2.Warnings, error)
	GetStack(guid string) (ccv2.Stack, ccv2.Warnings, error)
	GetStacks(queries ...ccv2.QQuery) ([]ccv2.Stack, ccv2.Warnings, error)
	GetStagingSpacesBySecurityGroup(securityGroupGUID string) ([]ccv2.Space, ccv2.Warnings, error)
	PollJob(job ccv2.Job) (ccv2.Warnings, error)
	RemoveSpaceFromRunningSecurityGroup(securityGroupGUID string, spaceGUID string) (ccv2.Warnings, error)
	RemoveSpaceFromStagingSecurityGroup(securityGroupGUID string, spaceGUID string) (ccv2.Warnings, error)
	ResourceMatch(resourcesToMatch []ccv2.Resource) ([]ccv2.Resource, ccv2.Warnings, error)
	RestageApplication(app ccv2.Application) (ccv2.Application, ccv2.Warnings, error)
	TargetCF(settings ccv2.TargetSettings) (ccv2.Warnings, error)
	UpdateApplication(app ccv2.Application) (ccv2.Application, ccv2.Warnings, error)
	UpdateRouteApplication(routeGUID string, appGUID string) (ccv2.Route, ccv2.Warnings, error)
	UploadApplicationPackage(appGUID string, existingResources []ccv2.Resource, newResources ccv2.Reader, newResourcesLength int64) (ccv2.Job, ccv2.Warnings, error)

	API() string
	APIVersion() string
	AuthorizationEndpoint() string
	DopplerEndpoint() string
	MinCLIVersion() string
	RoutingEndpoint() string
	TokenEndpoint() string
}
