package requirements

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
)

type FakeReqFactory struct {
	ApplicationName string
	Application     models.Application

	ServiceInstanceName string
	ServiceInstance     models.ServiceInstance

	ApplicationFails             bool
	LoginSuccess                 bool
	RoutingAPIEndpointSuccess    bool
	ApiEndpointSuccess           bool
	ValidAccessTokenSuccess      bool
	TargetedSpaceSuccess         bool
	TargetedOrgSuccess           bool
	BuildpackSuccess             bool
	SpaceRequirementFails        bool
	UserRequirementFails         bool
	OrganizationRequirementFails bool

	ServiceInstanceNotFound bool

	SpaceName string
	Space     models.Space

	OrganizationName   string
	Organization       models.Organization
	OrganizationFields models.OrganizationFields

	RouteHost   string
	RouteDomain string
	Route       models.Route

	DomainName string
	Domain     models.DomainFields

	UserUsername string
	UserFields   models.UserFields

	Buildpack models.Buildpack

	MinAPIVersionFeatureName     string
	MinAPIVersionRequiredVersion semver.Version
	MinAPIVersionSuccess         bool
}

func (f *FakeReqFactory) NewDEAApplicationRequirement(name string) requirements.DEAApplicationRequirement {
	return FakeRequirement{f, false}
}

func (f *FakeReqFactory) NewApplicationRequirement(name string) requirements.ApplicationRequirement {
	f.ApplicationName = name
	return FakeRequirement{f, !f.ApplicationFails}
}

func (f *FakeReqFactory) NewServiceInstanceRequirement(name string) requirements.ServiceInstanceRequirement {
	f.ServiceInstanceName = name
	return FakeRequirement{f, !f.ServiceInstanceNotFound}
}

func (f *FakeReqFactory) NewLoginRequirement() requirements.Requirement {
	return FakeRequirement{f, f.LoginSuccess}
}

func (f *FakeReqFactory) NewRoutingAPIRequirement() requirements.Requirement {
	return FakeRequirement{f, f.RoutingAPIEndpointSuccess}
}

func (f *FakeReqFactory) NewTargetedSpaceRequirement() requirements.Requirement {
	return FakeRequirement{f, f.TargetedSpaceSuccess}
}

func (f *FakeReqFactory) NewTargetedOrgRequirement() requirements.TargetedOrgRequirement {
	return FakeRequirement{f, f.TargetedOrgSuccess}
}

func (f *FakeReqFactory) NewSpaceRequirement(name string) requirements.SpaceRequirement {
	f.SpaceName = name
	return FakeRequirement{f, !f.SpaceRequirementFails}
}

func (f *FakeReqFactory) NewOrganizationRequirement(name string) requirements.OrganizationRequirement {
	f.OrganizationName = name
	return FakeRequirement{f, !f.OrganizationRequirementFails}
}

func (f *FakeReqFactory) NewDomainRequirement(name string) requirements.DomainRequirement {
	f.DomainName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewUserRequirement(username string, wantGuid bool) requirements.UserRequirement {
	f.UserUsername = username
	return FakeRequirement{f, !f.UserRequirementFails}
}

func (f *FakeReqFactory) NewBuildpackRequirement(buildpack string) requirements.BuildpackRequirement {
	f.Buildpack.Name = buildpack
	return FakeRequirement{f, f.BuildpackSuccess}
}

func (f *FakeReqFactory) NewApiEndpointRequirement() requirements.Requirement {
	return FakeRequirement{f, f.ApiEndpointSuccess}
}

func (f *FakeReqFactory) NewMinAPIVersionRequirement(featureName string, requiredVersion semver.Version) requirements.Requirement {
	f.MinAPIVersionFeatureName = featureName
	f.MinAPIVersionRequiredVersion = requiredVersion
	return FakeRequirement{f, f.MinAPIVersionSuccess}
}

type FakeRequirement struct {
	factory *FakeReqFactory
	success bool
}

func (r FakeRequirement) Execute() (success bool) {
	return r.success
}

func (r FakeRequirement) GetApplication() models.Application {
	return r.factory.Application
}

func (r FakeRequirement) GetServiceInstance() models.ServiceInstance {
	return r.factory.ServiceInstance
}

func (r FakeRequirement) SetSpaceName(name string) {
}

func (r FakeRequirement) GetSpace() models.Space {
	return r.factory.Space
}

func (r FakeRequirement) SetOrganizationName(name string) {
}

func (r FakeRequirement) GetOrganization() models.Organization {
	return r.factory.Organization
}

func (r FakeRequirement) GetOrganizationFields() models.OrganizationFields {
	return r.factory.OrganizationFields
}

func (r FakeRequirement) GetRoute() models.Route {
	return r.factory.Route
}

func (r FakeRequirement) GetDomain() models.DomainFields {
	return r.factory.Domain
}

func (r FakeRequirement) GetUser() models.UserFields {
	return r.factory.UserFields
}

func (r FakeRequirement) GetBuildpack() models.Buildpack {
	return r.factory.Buildpack
}
