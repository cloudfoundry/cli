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
	NewUserRequirement(username string) UserRequirement
}

type apiRequirementFactory struct {
	ui          terminal.UI
	config      *configuration.Configuration
	repoLocator api.RepositoryLocator
}

func NewFactory(ui terminal.UI, config *configuration.Configuration, repoLocator api.RepositoryLocator) (factory apiRequirementFactory) {
	return apiRequirementFactory{ui, config, repoLocator}
}

func (f apiRequirementFactory) NewApplicationRequirement(name string) ApplicationRequirement {
	return newApplicationRequirement(
		name,
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f apiRequirementFactory) NewServiceInstanceRequirement(name string) ServiceInstanceRequirement {
	return newServiceInstanceRequirement(
		name,
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f apiRequirementFactory) NewLoginRequirement() Requirement {
	return newLoginRequirement(
		f.ui,
		f.config,
	)
}
func (f apiRequirementFactory) NewValidAccessTokenRequirement() Requirement {
	return newValidAccessTokenRequirement(
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f apiRequirementFactory) NewSpaceRequirement(name string) SpaceRequirement {
	return newSpaceRequirement(
		name,
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f apiRequirementFactory) NewTargetedSpaceRequirement() Requirement {
	return newTargetedSpaceRequirement(
		f.ui,
		f.config,
	)
}

func (f apiRequirementFactory) NewTargetedOrgRequirement() TargetedOrgRequirement {
	return newTargetedOrgRequirement(
		f.ui,
		f.config,
	)
}

func (f apiRequirementFactory) NewOrganizationRequirement(name string) OrganizationRequirement {
	return newOrganizationRequirement(
		name,
		f.ui,
		f.repoLocator.GetOrganizationRepository(),
	)
}

func (f apiRequirementFactory) NewRouteRequirement(host, domain string) RouteRequirement {
	return newRouteRequirement(
		host,
		domain,
		f.ui,
		f.repoLocator.GetRouteRepository(),
	)
}

func (f apiRequirementFactory) NewDomainRequirement(name string) DomainRequirement {
	return newDomainRequirement(
		name,
		f.ui,
		f.repoLocator.GetDomainRepository(),
	)
}

func (f apiRequirementFactory) NewUserRequirement(username string) UserRequirement {
	return newUserRequirement(
		username,
		f.ui,
		f.repoLocator.GetUserRepository(),
	)
}
