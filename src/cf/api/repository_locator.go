package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
)

type RepositoryLocator struct {
	authRepo AuthenticationRepository

	endpointRepo      RemoteEndpointRepository
	organizationRepo  CloudControllerOrganizationRepository
	spaceRepo         CloudControllerSpaceRepository
	appRepo           CloudControllerApplicationRepository
	appBitsRepo       CloudControllerApplicationBitsRepository
	appSummaryRepo    CloudControllerAppSummaryRepository
	appEventsRepo     CloudControllerAppEventsRepository
	appFilesRepo      CloudControllerAppFilesRepository
	domainRepo        CloudControllerDomainRepository
	routeRepo         CloudControllerRouteRepository
	stackRepo         CloudControllerStackRepository
	serviceRepo       CloudControllerServiceRepository
	userRepo          CloudControllerUserRepository
	passwordRepo      CloudControllerPasswordRepository
	logsRepo          LoggregatorLogsRepository
	authTokenRepo     CloudControllerServiceAuthTokenRepository
	serviceBrokerRepo CloudControllerServiceBrokerRepository
}

func NewRepositoryLocator(config *configuration.Configuration, configRepo configuration.ConfigurationRepository, gatewaysByName map[string]net.Gateway) (loc RepositoryLocator) {
	authGateway := gatewaysByName["auth"]
	cloudControllerGateway := gatewaysByName["cloud-controller"]
	uaaGateway := gatewaysByName["uaa"]

	loc.authRepo = NewUAAAuthenticationRepository(authGateway, configRepo)

	// ensure gateway refreshers are set before passing them by value to repositories
	cloudControllerGateway.SetTokenRefresher(loc.authRepo)
	uaaGateway.SetTokenRefresher(loc.authRepo)

	loc.endpointRepo = NewEndpointRepository(config, cloudControllerGateway, configRepo)
	loc.organizationRepo = NewCloudControllerOrganizationRepository(config, cloudControllerGateway)
	loc.spaceRepo = NewCloudControllerSpaceRepository(config, cloudControllerGateway)
	loc.appRepo = NewCloudControllerApplicationRepository(config, cloudControllerGateway)
	loc.appBitsRepo = NewCloudControllerApplicationBitsRepository(config, cloudControllerGateway, cf.ApplicationZipper{})
	loc.appSummaryRepo = NewCloudControllerAppSummaryRepository(config, cloudControllerGateway, loc.appRepo)
	loc.appEventsRepo = NewCloudControllerAppEventsRepository(config, cloudControllerGateway)
	loc.appFilesRepo = NewCloudControllerAppFilesRepository(config, cloudControllerGateway)
	loc.domainRepo = NewCloudControllerDomainRepository(config, cloudControllerGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, cloudControllerGateway, loc.domainRepo)
	loc.stackRepo = NewCloudControllerStackRepository(config, cloudControllerGateway)
	loc.serviceRepo = NewCloudControllerServiceRepository(config, cloudControllerGateway)
	loc.userRepo = NewCloudControllerUserRepository(config, uaaGateway, cloudControllerGateway, loc.endpointRepo)
	loc.passwordRepo = NewCloudControllerPasswordRepository(config, uaaGateway, loc.endpointRepo)
	loc.logsRepo = NewLoggregatorLogsRepository(config, loc.endpointRepo)
	loc.authTokenRepo = NewCloudControllerServiceAuthTokenRepository(config, cloudControllerGateway)
	loc.serviceBrokerRepo = NewCloudControllerServiceBrokerRepository(config, cloudControllerGateway)

	return
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

func (locator RepositoryLocator) GetAppEventsRepository() AppEventsRepository {
	return locator.appEventsRepo
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

func (locator RepositoryLocator) GetUserRepository() UserRepository {
	return locator.userRepo
}

func (locator RepositoryLocator) GetPasswordRepository() PasswordRepository {
	return locator.passwordRepo
}

func (locator RepositoryLocator) GetLogsRepository() LogsRepository {
	return locator.logsRepo
}

func (locator RepositoryLocator) GetServiceAuthTokenRepository() ServiceAuthTokenRepository {
	return locator.authTokenRepo
}

func (locator RepositoryLocator) GetServiceBrokerRepository() ServiceBrokerRepository {
	return locator.serviceBrokerRepo
}
