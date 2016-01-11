package requirements

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"
)

//go:generate counterfeiter -o fakes/fake_factory.go . Factory
type Factory interface {
	NewApplicationRequirement(name string) ApplicationRequirement
	NewDEAApplicationRequirement(name string) DEAApplicationRequirement
	NewServiceInstanceRequirement(name string) ServiceInstanceRequirement
	NewLoginRequirement() Requirement
	NewRoutingAPIRequirement() Requirement
	NewSpaceRequirement(name string) SpaceRequirement
	NewTargetedSpaceRequirement() Requirement
	NewTargetedOrgRequirement() TargetedOrgRequirement
	NewOrganizationRequirement(name string) OrganizationRequirement
	NewDomainRequirement(name string) DomainRequirement
	NewUserRequirement(username string, wantGuid bool) UserRequirement
	NewBuildpackRequirement(buildpack string) BuildpackRequirement
	NewApiEndpointRequirement() Requirement
	NewMinAPIVersionRequirement(commandName string, requiredVersion semver.Version) Requirement
}

type apiRequirementFactory struct {
	ui          terminal.UI
	config      core_config.Reader
	repoLocator api.RepositoryLocator
}

func NewFactory(ui terminal.UI, config core_config.Reader, repoLocator api.RepositoryLocator) (factory apiRequirementFactory) {
	return apiRequirementFactory{ui, config, repoLocator}
}

func (f apiRequirementFactory) NewApplicationRequirement(name string) ApplicationRequirement {
	return NewApplicationRequirement(
		name,
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f apiRequirementFactory) NewDEAApplicationRequirement(name string) DEAApplicationRequirement {
	return NewDEAApplicationRequirement(
		name,
		f.ui,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f apiRequirementFactory) NewServiceInstanceRequirement(name string) ServiceInstanceRequirement {
	return NewServiceInstanceRequirement(
		name,
		f.ui,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f apiRequirementFactory) NewLoginRequirement() Requirement {
	return NewLoginRequirement(
		f.ui,
		f.config,
	)
}

func (f apiRequirementFactory) NewRoutingAPIRequirement() Requirement {
	return NewRoutingAPIRequirement(
		f.ui,
		f.config,
	)
}

func (f apiRequirementFactory) NewSpaceRequirement(name string) SpaceRequirement {
	return NewSpaceRequirement(
		name,
		f.ui,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f apiRequirementFactory) NewTargetedSpaceRequirement() Requirement {
	return NewTargetedSpaceRequirement(
		f.ui,
		f.config,
	)
}

func (f apiRequirementFactory) NewTargetedOrgRequirement() TargetedOrgRequirement {
	return NewTargetedOrgRequirement(
		f.ui,
		f.config,
	)
}

func (f apiRequirementFactory) NewOrganizationRequirement(name string) OrganizationRequirement {
	return NewOrganizationRequirement(
		name,
		f.ui,
		f.repoLocator.GetOrganizationRepository(),
	)
}

func (f apiRequirementFactory) NewDomainRequirement(name string) DomainRequirement {
	return NewDomainRequirement(
		name,
		f.ui,
		f.config,
		f.repoLocator.GetDomainRepository(),
	)
}

func (f apiRequirementFactory) NewUserRequirement(username string, wantGuid bool) UserRequirement {
	return NewUserRequirement(
		username,
		f.ui,
		f.repoLocator.GetUserRepository(),
		wantGuid,
	)
}

func (f apiRequirementFactory) NewBuildpackRequirement(buildpack string) BuildpackRequirement {
	return NewBuildpackRequirement(
		buildpack,
		f.ui,
		f.repoLocator.GetBuildpackRepository(),
	)
}

func (f apiRequirementFactory) NewApiEndpointRequirement() Requirement {
	return NewApiEndpointRequirement(
		f.ui,
		f.config,
	)
}

func (f apiRequirementFactory) NewMinAPIVersionRequirement(commandName string, requiredVersion semver.Version) Requirement {
	return NewMinAPIVersionRequirement(
		f.ui,
		f.config,
		commandName,
		requiredVersion,
	)
}
