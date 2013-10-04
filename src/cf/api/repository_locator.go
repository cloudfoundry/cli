package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
)

type RepositoryLocator struct {
	authRepo AuthenticationRepository

	configurationRepo configuration.ConfigurationDiskRepository
	endpointRepo      RemoteEndpointRepository
	organizationRepo  CloudControllerOrganizationRepository
	spaceRepo         CloudControllerSpaceRepository
	appRepo           CloudControllerApplicationRepository
	appBitsRepo       CloudControllerApplicationBitsRepository
	appSummaryRepo    CloudControllerAppSummaryRepository
	appFilesRepo      CloudControllerAppFilesRepository
	domainRepo        CloudControllerDomainRepository
	routeRepo         CloudControllerRouteRepository
	stackRepo         CloudControllerStackRepository
	serviceRepo       CloudControllerServiceRepository
	passwordRepo      CloudControllerPasswordRepository
	logsRepo          LoggregatorLogsRepository
}

func NewRepositoryLocator(config *configuration.Configuration) (loc RepositoryLocator) {
	authGateway := net.NewUAAAuthGateway()

	loc.configurationRepo = configuration.NewConfigurationDiskRepository()
	loc.authRepo = NewUAAAuthenticationRepository(authGateway, loc.configurationRepo)

	cloudControllerGateway := net.NewCloudControllerGateway(loc.authRepo)
	uaaGateway := net.NewUAAGateway(loc.authRepo)

	loc.endpointRepo = NewEndpointRepository(config, cloudControllerGateway, loc.configurationRepo)
	loc.organizationRepo = NewCloudControllerOrganizationRepository(config, cloudControllerGateway)
	loc.spaceRepo = NewCloudControllerSpaceRepository(config, cloudControllerGateway)
	loc.appRepo = NewCloudControllerApplicationRepository(config, cloudControllerGateway)
	loc.appBitsRepo = NewCloudControllerApplicationBitsRepository(config, cloudControllerGateway, cf.ApplicationZipper{})
	loc.appSummaryRepo = NewCloudControllerAppSummaryRepository(config, cloudControllerGateway, loc.appRepo)
	loc.appFilesRepo = NewCloudControllerAppFilesRepository(config, cloudControllerGateway)
	loc.domainRepo = NewCloudControllerDomainRepository(config, cloudControllerGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, cloudControllerGateway, loc.domainRepo)
	loc.stackRepo = NewCloudControllerStackRepository(config, cloudControllerGateway)
	loc.serviceRepo = NewCloudControllerServiceRepository(config, cloudControllerGateway)
	loc.passwordRepo = NewCloudControllerPasswordRepository(config, uaaGateway)
	loc.logsRepo = NewLoggregatorLogsRepository(config, cloudControllerGateway, LoggregatorHost)

	return
}

func (locator RepositoryLocator) GetConfigurationRepository() configuration.ConfigurationRepository {
	return locator.configurationRepo
}

func (locator RepositoryLocator) GetAuthenticationRepository() AuthenticationRepository {
	return locator.authRepo
}

func (locator RepositoryLocator) GetEndpointRepository() EndpointRepository {
	return locator.endpointRepo
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

func (locator RepositoryLocator) GetApplicationBitsRepository() ApplicationBitsRepository {
	return locator.appBitsRepo
}

func (locator RepositoryLocator) GetAppSummaryRepository() AppSummaryRepository {
	return locator.appSummaryRepo
}

func (locator RepositoryLocator) GetAppFilesRepository() AppFilesRepository {
	return locator.appFilesRepo
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

func (locator RepositoryLocator) GetLogsRepository() LogsRepository {
	return locator.logsRepo
}
