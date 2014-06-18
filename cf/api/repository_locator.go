package api

import (
	"crypto/tls"
	"github.com/cloudfoundry/cli/cf/api/strategy"
	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
	consumer "github.com/cloudfoundry/loggregator_consumer"
)

type RepositoryLocator struct {
	authRepo                        AuthenticationRepository
	curlRepo                        CurlRepository
	endpointRepo                    RemoteEndpointRepository
	organizationRepo                CloudControllerOrganizationRepository
	quotaRepo                       CloudControllerQuotaRepository
	spaceRepo                       CloudControllerSpaceRepository
	appRepo                         CloudControllerApplicationRepository
	appBitsRepo                     CloudControllerApplicationBitsRepository
	appSummaryRepo                  CloudControllerAppSummaryRepository
	appInstancesRepo                CloudControllerAppInstancesRepository
	appEventsRepo                   CloudControllerAppEventsRepository
	appFilesRepo                    CloudControllerAppFilesRepository
	domainRepo                      CloudControllerDomainRepository
	routeRepo                       CloudControllerRouteRepository
	stackRepo                       CloudControllerStackRepository
	serviceRepo                     CloudControllerServiceRepository
	serviceBindingRepo              CloudControllerServiceBindingRepository
	serviceSummaryRepo              CloudControllerServiceSummaryRepository
	userRepo                        CloudControllerUserRepository
	passwordRepo                    CloudControllerPasswordRepository
	logsRepo                        LogsRepository
	authTokenRepo                   CloudControllerServiceAuthTokenRepository
	serviceBrokerRepo               CloudControllerServiceBrokerRepository
	userProvidedServiceInstanceRepo CCUserProvidedServiceInstanceRepository
	buildpackRepo                   CloudControllerBuildpackRepository
	buildpackBitsRepo               CloudControllerBuildpackBitsRepository
	appSecurityGroupRepo            ApplicationSecurityGroupRepo
}

func NewRepositoryLocator(config configuration.ReadWriter, gatewaysByName map[string]net.Gateway) (loc RepositoryLocator) {
	strategy := strategy.NewEndpointStrategy(config.ApiVersion())

	authGateway := gatewaysByName["auth"]
	cloudControllerGateway := gatewaysByName["cloud-controller"]
	uaaGateway := gatewaysByName["uaa"]
	loc.authRepo = NewUAAAuthenticationRepository(authGateway, config)

	// ensure gateway refreshers are set before passing them by value to repositories
	cloudControllerGateway.SetTokenRefresher(loc.authRepo)
	uaaGateway.SetTokenRefresher(loc.authRepo)

	tlsConfig := net.NewTLSConfig([]tls.Certificate{}, config.IsSSLDisabled())
	loggregatorConsumer := consumer.New(config.LoggregatorEndpoint(), tlsConfig, nil)

	loc.appBitsRepo = NewCloudControllerApplicationBitsRepository(config, cloudControllerGateway, app_files.ApplicationZipper{})
	loc.appEventsRepo = NewCloudControllerAppEventsRepository(config, cloudControllerGateway, strategy)
	loc.appFilesRepo = NewCloudControllerAppFilesRepository(config, cloudControllerGateway)
	loc.appRepo = NewCloudControllerApplicationRepository(config, cloudControllerGateway)
	loc.appSummaryRepo = NewCloudControllerAppSummaryRepository(config, cloudControllerGateway)
	loc.appInstancesRepo = NewCloudControllerAppInstancesRepository(config, cloudControllerGateway)
	loc.authTokenRepo = NewCloudControllerServiceAuthTokenRepository(config, cloudControllerGateway)
	loc.curlRepo = NewCloudControllerCurlRepository(config, cloudControllerGateway)
	loc.domainRepo = NewCloudControllerDomainRepository(config, cloudControllerGateway, strategy)
	loc.endpointRepo = NewEndpointRepository(config, cloudControllerGateway)
	loc.logsRepo = NewLoggregatorLogsRepository(config, loggregatorConsumer, loc.authRepo)
	loc.organizationRepo = NewCloudControllerOrganizationRepository(config, cloudControllerGateway)
	loc.passwordRepo = NewCloudControllerPasswordRepository(config, uaaGateway)
	loc.quotaRepo = NewCloudControllerQuotaRepository(config, cloudControllerGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, cloudControllerGateway)
	loc.stackRepo = NewCloudControllerStackRepository(config, cloudControllerGateway)
	loc.serviceRepo = NewCloudControllerServiceRepository(config, cloudControllerGateway)
	loc.serviceBindingRepo = NewCloudControllerServiceBindingRepository(config, cloudControllerGateway)
	loc.serviceBrokerRepo = NewCloudControllerServiceBrokerRepository(config, cloudControllerGateway)
	loc.serviceSummaryRepo = NewCloudControllerServiceSummaryRepository(config, cloudControllerGateway)
	loc.spaceRepo = NewCloudControllerSpaceRepository(config, cloudControllerGateway)
	loc.userProvidedServiceInstanceRepo = NewCCUserProvidedServiceInstanceRepository(config, cloudControllerGateway)
	loc.userRepo = NewCloudControllerUserRepository(config, uaaGateway, cloudControllerGateway)
	loc.buildpackRepo = NewCloudControllerBuildpackRepository(config, cloudControllerGateway)
	loc.buildpackBitsRepo = NewCloudControllerBuildpackBitsRepository(config, cloudControllerGateway, app_files.ApplicationZipper{})
	loc.appSecurityGroupRepo = NewApplicationSecurityGroupRepo(config, cloudControllerGateway)

	return
}

func (locator RepositoryLocator) GetAuthenticationRepository() AuthenticationRepository {
	return locator.authRepo
}

func (locator RepositoryLocator) GetCurlRepository() CurlRepository {
	return locator.curlRepo
}

func (locator RepositoryLocator) GetEndpointRepository() EndpointRepository {
	return locator.endpointRepo
}

func (locator RepositoryLocator) GetOrganizationRepository() OrganizationRepository {
	return locator.organizationRepo
}

func (locator RepositoryLocator) GetQuotaRepository() QuotaRepository {
	return locator.quotaRepo
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

func (locator RepositoryLocator) GetAppInstancesRepository() AppInstancesRepository {
	return locator.appInstancesRepo
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

func (locator RepositoryLocator) GetServiceBindingRepository() ServiceBindingRepository {
	return locator.serviceBindingRepo
}

func (locator RepositoryLocator) GetServiceSummaryRepository() ServiceSummaryRepository {
	return locator.serviceSummaryRepo
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

func (locator RepositoryLocator) GetUserProvidedServiceInstanceRepository() UserProvidedServiceInstanceRepository {
	return locator.userProvidedServiceInstanceRepo
}

func (locator RepositoryLocator) GetBuildpackRepository() BuildpackRepository {
	return locator.buildpackRepo
}

func (locator RepositoryLocator) GetBuildpackBitsRepository() BuildpackBitsRepository {
	return locator.buildpackBitsRepo
}

func (locator RepositoryLocator) GetApplicationSecurityGroupRepository() ApplicationSecurityGroupRepo {
	return locator.appSecurityGroupRepo
}
