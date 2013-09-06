package api

import (
	"cf/configuration"
)

type RepositoryLocator struct {
	config *configuration.Configuration

	configurationRepo configuration.ConfigurationDiskRepository
	organizationRepo  CloudControllerOrganizationRepository
	spaceRepo         CloudControllerSpaceRepository
	appRepo           CloudControllerApplicationRepository
	domainRepo        CloudControllerDomainRepository
	routeRepo         CloudControllerRouteRepository
	stackRepo         CloudControllerStackRepository
	serviceRepo       CloudControllerServiceRepository
}

func NewRepositoryLocator(config *configuration.Configuration) (locator RepositoryLocator) {
	locator.config = config

	locator.configurationRepo = configuration.NewConfigurationDiskRepository()
	locator.organizationRepo = NewCloudControllerOrganizationRepository(config)
	locator.spaceRepo = CloudControllerSpaceRepository{}
	locator.appRepo = NewCloudControllerApplicationRepository(config)
	locator.domainRepo = NewCloudControllerDomainRepository(config)
	locator.routeRepo = NewCloudControllerRouteRepository(config)
	locator.stackRepo = CloudControllerStackRepository{}
	locator.serviceRepo = CloudControllerServiceRepository{}

	return
}

func (locator RepositoryLocator) GetConfig() *configuration.Configuration {
	return locator.config
}

func (locator RepositoryLocator) GetConfigurationRepository() configuration.ConfigurationRepository {
	return locator.configurationRepo
}

func (locator RepositoryLocator) GetOrganizationRepository() OrganizationRepository {
	return locator.organizationRepo
}

func (locator RepositoryLocator) GetSpaceRepository() SpaceRepository {
	return locator.spaceRepo
}

func (locator RepositoryLocator) GetApplicationRepository() ApplicationRepository {
	return locator.appRepo
}

func (locator RepositoryLocator) GetDomainRepository() DomainRepository {
	return locator.domainRepo
}

func (locator RepositoryLocator) GetRouteRepository() RouteRepository {
	return locator.routeRepo
}

func (locator RepositoryLocator) GetStackRepository() StackRepository {
	return locator.stackRepo
}

func (locator RepositoryLocator) GetServiceRepository() ServiceRepository {
	return locator.serviceRepo
}
