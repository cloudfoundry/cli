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
	passwordRepo      CloudControllerPasswordRepository
}

func NewRepositoryLocator(config *configuration.Configuration) (locator RepositoryLocator) {
	locator.config = config
	locator.configurationRepo = configuration.NewConfigurationDiskRepository()

	authenticator := NewUAAAuthenticator(locator.configurationRepo)
	apiClient := NewApiClient(authenticator)

	locator.organizationRepo = NewCloudControllerOrganizationRepository(config, apiClient)
	locator.spaceRepo = NewCloudControllerSpaceRepository(config, apiClient)
	locator.appRepo = NewCloudControllerApplicationRepository(config, apiClient)
	locator.domainRepo = NewCloudControllerDomainRepository(config, apiClient)
	locator.routeRepo = NewCloudControllerRouteRepository(config, apiClient)
	locator.stackRepo = NewCloudControllerStackRepository(config, apiClient)
	locator.serviceRepo = NewCloudControllerServiceRepository(config, apiClient)
	locator.passwordRepo = NewCloudControllerPasswordRepository(config, apiClient)

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

func (locator RepositoryLocator) GetPasswordRepository() PasswordRepository {
	return locator.passwordRepo
}
