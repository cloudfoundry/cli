package api

import (
	"cf/configuration"
	"cf/net"
)

type RepositoryLocator struct {
	config *configuration.Configuration

	authenticator          Authenticator
	cloudControllerGateway net.Gateway
	uaaGateway             net.Gateway

	configurationRepo configuration.ConfigurationDiskRepository
	organizationRepo  CloudControllerOrganizationRepository
	spaceRepo         CloudControllerSpaceRepository
	appRepo           CloudControllerApplicationRepository
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
	loc.config = config
	loc.configurationRepo = configuration.NewConfigurationDiskRepository()

	authGateway := net.NewUAAAuthGateway()
	loc.authenticator = NewUAAAuthenticator(authGateway, loc.configurationRepo)

	loc.cloudControllerGateway = net.NewCloudControllerGateway(loc.authenticator)
	loc.uaaGateway = net.NewUAAGateway(loc.authenticator)

	loc.organizationRepo = NewCloudControllerOrganizationRepository(config, loc.cloudControllerGateway)
	loc.spaceRepo = NewCloudControllerSpaceRepository(config, loc.cloudControllerGateway)
	loc.appRepo = NewCloudControllerApplicationRepository(config, loc.cloudControllerGateway)
	loc.appSummaryRepo = NewCloudControllerAppSummaryRepository(config, loc.cloudControllerGateway, loc.appRepo)
	loc.appFilesRepo = NewCloudControllerAppFilesRepository(config, loc.cloudControllerGateway)
	loc.domainRepo = NewCloudControllerDomainRepository(config, loc.cloudControllerGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, loc.cloudControllerGateway)
	loc.stackRepo = NewCloudControllerStackRepository(config, loc.cloudControllerGateway)
	loc.serviceRepo = NewCloudControllerServiceRepository(config, loc.cloudControllerGateway)
	loc.passwordRepo = NewCloudControllerPasswordRepository(config, loc.uaaGateway)
	loc.logsRepo = NewLoggregatorLogsRepository(config, loc.cloudControllerGateway, LoggregatorHost)

	return
}

func (locator RepositoryLocator) GetConfig() *configuration.Configuration {
	return locator.config
}

func (locator RepositoryLocator) GetConfigurationRepository() configuration.ConfigurationRepository {
	return locator.configurationRepo
}

func (locator RepositoryLocator) GetAuthenticator() Authenticator {
	return locator.authenticator
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
