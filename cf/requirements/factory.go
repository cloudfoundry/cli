package requirements

import (
	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"github.com/blang/semver/v4"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Factory

type Factory interface {
	NewApplicationRequirement(name string) ApplicationRequirement
	NewDEAApplicationRequirement(name string) DEAApplicationRequirement
	NewServiceInstanceRequirement(name string) ServiceInstanceRequirement
	NewLoginRequirement() Requirement
	NewSpaceRequirement(name string) SpaceRequirement
	NewTargetedSpaceRequirement() Requirement
	NewTargetedOrgRequirement() TargetedOrgRequirement
	NewOrganizationRequirement(name string) OrganizationRequirement
	NewDomainRequirement(name string) DomainRequirement
	NewUserRequirement(username string, wantGUID bool) UserRequirement
	NewClientRequirement(username string) UserRequirement
	NewBuildpackRequirement(buildpack, stack string) BuildpackRequirement
	NewAPIEndpointRequirement() Requirement
	NewMinAPIVersionRequirement(commandName string, requiredVersion semver.Version) Requirement
	NewMaxAPIVersionRequirement(commandName string, maximumVersion semver.Version) Requirement
	NewUnsupportedLegacyFlagRequirement(flags ...string) Requirement
	NewUsageRequirement(Usable, string, func() bool) Requirement
	NewNumberArguments([]string, ...string) Requirement
}

type apiRequirementFactory struct {
	config      coreconfig.ReadWriter
	repoLocator api.RepositoryLocator
}

func NewFactory(config coreconfig.ReadWriter, repoLocator api.RepositoryLocator) (factory apiRequirementFactory) {
	return apiRequirementFactory{config, repoLocator}
}

func (f apiRequirementFactory) NewApplicationRequirement(name string) ApplicationRequirement {
	return NewApplicationRequirement(
		name,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f apiRequirementFactory) NewDEAApplicationRequirement(name string) DEAApplicationRequirement {
	return NewDEAApplicationRequirement(
		name,
		f.repoLocator.GetApplicationRepository(),
	)
}

func (f apiRequirementFactory) NewServiceInstanceRequirement(name string) ServiceInstanceRequirement {
	return NewServiceInstanceRequirement(
		name,
		f.repoLocator.GetServiceRepository(),
	)
}

func (f apiRequirementFactory) NewLoginRequirement() Requirement {
	return NewLoginRequirement(
		f.config,
	)
}

func (f apiRequirementFactory) NewSpaceRequirement(name string) SpaceRequirement {
	return NewSpaceRequirement(
		name,
		f.repoLocator.GetSpaceRepository(),
	)
}

func (f apiRequirementFactory) NewTargetedSpaceRequirement() Requirement {
	return NewTargetedSpaceRequirement(
		f.config,
	)
}

func (f apiRequirementFactory) NewTargetedOrgRequirement() TargetedOrgRequirement {
	return NewTargetedOrgRequirement(
		f.config,
	)
}

func (f apiRequirementFactory) NewOrganizationRequirement(name string) OrganizationRequirement {
	return NewOrganizationRequirement(
		name,
		f.repoLocator.GetOrganizationRepository(),
	)
}

func (f apiRequirementFactory) NewDomainRequirement(name string) DomainRequirement {
	return NewDomainRequirement(
		name,
		f.config,
		f.repoLocator.GetDomainRepository(),
	)
}

func (f apiRequirementFactory) NewUserRequirement(username string, wantGUID bool) UserRequirement {
	return NewUserRequirement(
		username,
		f.repoLocator.GetUserRepository(),
		wantGUID,
	)
}

func (f apiRequirementFactory) NewClientRequirement(username string) UserRequirement {
	return NewClientRequirement(
		username,
		f.repoLocator.GetClientRepository(),
	)
}

func (f apiRequirementFactory) NewBuildpackRequirement(buildpack, stack string) BuildpackRequirement {
	return NewBuildpackRequirement(
		buildpack,
		stack,
		f.repoLocator.GetBuildpackRepository(),
	)
}

func (f apiRequirementFactory) NewAPIEndpointRequirement() Requirement {
	return NewAPIEndpointRequirement(
		f.config,
	)
}

func (f apiRequirementFactory) NewMinAPIVersionRequirement(commandName string, requiredVersion semver.Version) Requirement {
	r := NewMinAPIVersionRequirement(
		f.config,
		commandName,
		requiredVersion,
	)

	refresher := coreconfig.APIConfigRefresher{
		Endpoint:     f.config.APIEndpoint(),
		EndpointRepo: f.repoLocator.GetEndpointRepository(),
		Config:       f.config,
	}

	return NewConfigRefreshingRequirement(r, refresher)
}

func (f apiRequirementFactory) NewMaxAPIVersionRequirement(commandName string, maximumVersion semver.Version) Requirement {
	return NewMaxAPIVersionRequirement(
		f.config,
		commandName,
		maximumVersion,
	)
}

func (f apiRequirementFactory) NewUnsupportedLegacyFlagRequirement(flags ...string) Requirement {
	return NewUnsupportedLegacyFlagRequirement(flags)
}

func (f apiRequirementFactory) NewUsageRequirement(cmd Usable, errorMessage string, pred func() bool) Requirement {
	return NewUsageRequirement(cmd, errorMessage, pred)
}

func (f apiRequirementFactory) NewNumberArguments(passedArgs []string, expectedArgs ...string) Requirement {
	return NewNumberArguments(passedArgs, expectedArgs)
}
