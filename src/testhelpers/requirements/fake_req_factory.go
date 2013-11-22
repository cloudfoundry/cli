package requirements

import (
	"cf"
	"cf/requirements"
)

type FakeReqFactory struct {
	ApplicationName string
	Application     cf.Application

	ServiceInstanceName string
	ServiceInstance     cf.ServiceInstance

	LoginSuccess            bool
	ValidAccessTokenSuccess bool
	TargetedSpaceSuccess    bool
	TargetedOrgSuccess      bool
	BuildpackSuccess		bool

	SpaceName string
	Space     cf.Space

	OrganizationName string
	Organization cf.Organization
	OrganizationFields cf.OrganizationFields

	RouteHost   string
	RouteDomain string
	Route       cf.Route

	DomainName string
	Domain     cf.Domain

	UserUsername string
	UserFields         cf.UserFields

	Buildpack     cf.Buildpack
}

func (f *FakeReqFactory) NewApplicationRequirement(name string) requirements.ApplicationRequirement {
	f.ApplicationName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewServiceInstanceRequirement(name string) requirements.ServiceInstanceRequirement {
	f.ServiceInstanceName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewLoginRequirement() requirements.Requirement {
	return FakeRequirement{f, f.LoginSuccess}
}

func (f *FakeReqFactory) NewValidAccessTokenRequirement() requirements.Requirement {
	return FakeRequirement{f, f.ValidAccessTokenSuccess}
}

func (f *FakeReqFactory) NewTargetedSpaceRequirement() requirements.Requirement {
	return FakeRequirement{f, f.TargetedSpaceSuccess}
}

func (f *FakeReqFactory) NewTargetedOrgRequirement() requirements.TargetedOrgRequirement {
	return FakeRequirement{f, f.TargetedOrgSuccess}
}

func (f *FakeReqFactory) NewSpaceRequirement(name string) requirements.SpaceRequirement {
	f.SpaceName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewOrganizationRequirement(name string) requirements.OrganizationRequirement {
	f.OrganizationName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewDomainRequirement(name string) requirements.DomainRequirement {
	f.DomainName = name
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewUserRequirement(username string) requirements.UserRequirement {
	f.UserUsername = username
	return FakeRequirement{f, true}
}

func (f *FakeReqFactory) NewBuildpackRequirement(buildpack string) requirements.BuildpackRequirement {
	f.Buildpack.Name = buildpack
	return FakeRequirement{f, f.BuildpackSuccess}
}

type FakeRequirement struct {
	factory *FakeReqFactory
	success bool
}

func (r FakeRequirement) Execute() (success bool) {
	return r.success
}

func (r FakeRequirement) GetApplication() cf.Application {
	return r.factory.Application
}

func (r FakeRequirement) GetServiceInstance() cf.ServiceInstance {
	return r.factory.ServiceInstance
}

func (r FakeRequirement) GetSpace() cf.Space {
	return r.factory.Space
}

func (r FakeRequirement) GetOrganization() cf.Organization {
	return r.factory.Organization
}

func (r FakeRequirement) GetOrganizationFields() cf.OrganizationFields {
	return r.factory.OrganizationFields
}

func (r FakeRequirement) GetRoute() cf.Route {
	return r.factory.Route
}

func (r FakeRequirement) GetDomain() cf.Domain {
	return r.factory.Domain
}

func (r FakeRequirement) GetUser() cf.UserFields {
	return r.factory.UserFields
}

func (r FakeRequirement) GetBuildpack() cf.Buildpack {
	return r.factory.Buildpack
}
