package requirements

import (
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type FakeReqFactory struct {
	ApplicationName string
	Application     models.Application

	ServiceInstanceName string
	ServiceInstance     models.ServiceInstance

	ApplicationFails             bool
	LoginSuccess                 bool
	RoutingAPIEndpointSuccess    bool
	RoutingAPIEndpointPanic      bool
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

	MinCCApiVersionCommandName string
	MinCCApiVersionMajor       int
	MinCCApiVersionMinor       int
	MinCCApiVersionPatch       int

	UI terminal.UI
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
	return FakePanicRequirement{FakeRequirement{f, f.RoutingAPIEndpointSuccess}, f.RoutingAPIEndpointPanic, "Routing API URI missing. Please log in again to set the URI automatically.", f.UI}
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

func (f *FakeReqFactory) NewMinCCApiVersionRequirement(commandName string, major, minor, patch int) requirements.Requirement {
	f.MinCCApiVersionCommandName = commandName
	f.MinCCApiVersionMajor = major
	f.MinCCApiVersionMinor = minor
	f.MinCCApiVersionPatch = patch
	return FakeRequirement{f, true}
}

type FakeRequirement struct {
	factory *FakeReqFactory
	success bool
}

type FakePanicRequirement struct {
	FakeRequirement
	panicFlag    bool
	panicMessage string
	ui           terminal.UI
}

func (r FakeRequirement) Execute() (success bool) {
	return r.success
}

func (r FakePanicRequirement) Execute() (success bool) {
	if r.panicFlag {
		r.ui.Say(T("FAILED"))
		r.ui.Say(T(r.panicMessage))
		panic("")
	}
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
