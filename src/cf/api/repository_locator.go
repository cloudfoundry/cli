package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
)

type RepositoryLocator struct {
	authRepo               AuthenticationRepository
	cloudControllerGateway net.Gateway
	uaaGateway             net.Gateway

	configurationRepo configuration.ConfigurationDiskRepository
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

	loc.cloudControllerGateway = net.NewCloudControllerGateway(loc.authRepo)
	loc.uaaGateway = net.NewUAAGateway(loc.authRepo)

	loc.organizationRepo = NewCloudControllerOrganizationRepository(config, loc.cloudControllerGateway)
	loc.spaceRepo = NewCloudControllerSpaceRepository(config, loc.cloudControllerGateway)
	loc.appRepo = NewCloudControllerApplicationRepository(config, loc.cloudControllerGateway)
	loc.appBitsRepo = NewCloudControllerApplicationBitsRepository(config, loc.cloudControllerGateway, cf.ApplicationZipper{})
	loc.appSummaryRepo = NewCloudControllerAppSummaryRepository(config, loc.cloudControllerGateway, loc.appRepo)
	loc.appFilesRepo = NewCloudControllerAppFilesRepository(config, loc.cloudControllerGateway)
	loc.domainRepo = NewCloudControllerDomainRepository(config, loc.cloudControllerGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, loc.cloudControllerGateway, loc.domainRepo)
	loc.stackRepo = NewCloudControllerStackRepository(config, loc.cloudControllerGateway)
	loc.serviceRepo = NewCloudControllerServiceRepository(config, loc.cloudControllerGateway)
	loc.passwordRepo = NewCloudControllerPasswordRepository(config, loc.uaaGateway)
	loc.logsRepo = NewLoggregatorLogsRepository(config, loc.cloudControllerGateway, LoggregatorHost)

	return
}

func (locator RepositoryLocator) GetConfigurationRepository() configuration.ConfigurationRepository {
	return locator.configurationRepo
}

func (locator RepositoryLocator) GetAuthenticationRepository() AuthenticationRepository {
	return locator.authRepo
}

func (locator RepositoryLocator) GetCloudControllerGateway() net.Gateway {
	return locator.cloudControllerGateway
}

func (locator RepositoryLocator) GetUAAGateway() net.Gateway {
	return locator.uaaGateway
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
