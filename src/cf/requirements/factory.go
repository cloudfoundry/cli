package requirements

import (
	"cf/api"
	"cf/configuration"
	"cf/terminal"
)

type Requirement interface {
	Execute() (success bool)
}

type Factory interface {
	NewApplicationRequirement(name string) ApplicationRequirement
	NewServiceInstanceRequirement(name string) ServiceInstanceRequirement
	NewLoginRequirement() Requirement
	NewValidAccessTokenRequirement() Requirement
	NewSpaceRequirement(name string) SpaceRequirement
	NewTargetedSpaceRequirement() Requirement
	NewTargetedOrgRequirement() TargetedOrgRequirement
	NewOrganizationRequirement(name string) OrganizationRequirement
	NewRouteRequirement(host, domain string) RouteRequirement
	NewDomainRequirement(name string) DomainRequirement
}

type ApiRequirementFactory struct {
	ui          terminal.UI
	config      *configuration.Configuration
	repoLocator api.RepositoryLocator
}

func NewFactory(ui terminal.UI, config *configuration.Configuration, repoLocator api.RepositoryLocator) (factory ApiRequirementFactory) {
	return ApiRequirementFactory{ui, config, repoLocator}
}

func (f ApiRequirementFactory) NewApplicationRequirement(name string) ApplicationRequirement {
	return NewApplicationRequirement(
		name,
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f ApiRequirementFactory) NewServiceInstanceRequirement(name string) ServiceInstanceRequirement {
	return NewServiceInstanceRequirement(
		name,
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f ApiRequirementFactory) NewLoginRequirement() Requirement {
	return NewLoginRequirement(
		f.ui,
		f.config,
	)
}
func (f ApiRequirementFactory) NewValidAccessTokenRequirement() Requirement {
	return NewValidAccessTokenRequirement(
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f ApiRequirementFactory) NewSpaceRequirement(name string) SpaceRequirement {
	return NewSpaceRequirement(
		name,
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f ApiRequirementFactory) NewTargetedSpaceRequirement() Requirement {
	return NewTargetedSpaceRequirement(
		f.ui,
		f.config,
	)
}

func (f ApiRequirementFactory) NewTargetedOrgRequirement() TargetedOrgRequirement {
	return NewTargetedOrgRequirement(
		f.ui,
		f.config,
	)
}

func (f ApiRequirementFactory) NewOrganizationRequirement(name string) OrganizationRequirement {
	return NewOrganizationRequirement(
		name,
		f.ui,
		f.repoLocator.GetOrganizationRepository(),
	)
}

func (f ApiRequirementFactory) NewRouteRequirement(host, domain string) RouteRequirement {
	return NewRouteRequirement(
		host,
		domain,
		f.ui,
		f.repoLocator.GetRouteRepository(),
	)
}

func (f ApiRequirementFactory) NewDomainRequirement(name string) DomainRequirement {
	return NewDomainRequirement(
		name,
		f.ui,
		f.repoLocator.GetDomainRepository(),
	)
}
