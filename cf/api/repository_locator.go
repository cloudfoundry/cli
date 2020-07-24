package api

import (
	"code.cloudfoundry.org/cli/cf/api/appinstances"
	"code.cloudfoundry.org/cli/cf/api/applications"
	"code.cloudfoundry.org/cli/cf/api/authentication"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/api/password"
	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/trace"
)

type RepositoryLocator struct {
	authRepo                        authentication.Repository
	curlRepo                        CurlRepository
	endpointRepo                    coreconfig.EndpointRepository
	organizationRepo                organizations.OrganizationRepository
	spaceRepo                       spaces.SpaceRepository
	appRepo                         applications.Repository
	appSummaryRepo                  AppSummaryRepository
	appInstancesRepo                appinstances.Repository
	domainRepo                      DomainRepository
	routeRepo                       RouteRepository
	routingAPIRepo                  RoutingAPIRepository
	serviceRepo                     ServiceRepository
	serviceKeyRepo                  ServiceKeyRepository
	serviceBindingRepo              ServiceBindingRepository
	routeServiceBindingRepo         RouteServiceBindingRepository
	serviceSummaryRepo              ServiceSummaryRepository
	userRepo                        UserRepository
	clientRepo                      ClientRepository
	passwordRepo                    password.Repository
	authTokenRepo                   ServiceAuthTokenRepository
	serviceBrokerRepo               ServiceBrokerRepository
	servicePlanRepo                 CloudControllerServicePlanRepository
	servicePlanVisibilityRepo       ServicePlanVisibilityRepository
	userProvidedServiceInstanceRepo UserProvidedServiceInstanceRepository
	buildpackRepo                   BuildpackRepository
}

func NewRepositoryLocator(config coreconfig.ReadWriter, gatewaysByName map[string]net.Gateway, logger trace.Printer, envDialTimeout string) (loc RepositoryLocator) {
	cloudControllerGateway := gatewaysByName["cloud-controller"]
	routingAPIGateway := gatewaysByName["routing-api"]
	uaaGateway := gatewaysByName["uaa"]
	loc.authRepo = authentication.NewUAARepository(uaaGateway, config, net.NewRequestDumper(logger))

	// ensure gateway refreshers are set before passing them by value to repositories
	cloudControllerGateway.SetTokenRefresher(loc.authRepo)
	uaaGateway.SetTokenRefresher(loc.authRepo)

	loc.appRepo = applications.NewCloudControllerRepository(config, cloudControllerGateway)
	loc.appSummaryRepo = NewCloudControllerAppSummaryRepository(config, cloudControllerGateway)
	loc.appInstancesRepo = appinstances.NewCloudControllerAppInstancesRepository(config, cloudControllerGateway)
	loc.authTokenRepo = NewCloudControllerServiceAuthTokenRepository(config, cloudControllerGateway)
	loc.curlRepo = NewCloudControllerCurlRepository(config, cloudControllerGateway)
	loc.domainRepo = NewCloudControllerDomainRepository(config, cloudControllerGateway)
	loc.endpointRepo = NewEndpointRepository(cloudControllerGateway)

	loc.organizationRepo = organizations.NewCloudControllerOrganizationRepository(config, cloudControllerGateway)
	loc.passwordRepo = password.NewCloudControllerRepository(config, uaaGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, cloudControllerGateway)
	loc.routeServiceBindingRepo = NewCloudControllerRouteServiceBindingRepository(config, cloudControllerGateway)
	loc.routingAPIRepo = NewRoutingAPIRepository(config, routingAPIGateway)
	loc.serviceRepo = NewCloudControllerServiceRepository(config, cloudControllerGateway)
	loc.serviceKeyRepo = NewCloudControllerServiceKeyRepository(config, cloudControllerGateway)
	loc.serviceBindingRepo = NewCloudControllerServiceBindingRepository(config, cloudControllerGateway)
	loc.serviceBrokerRepo = NewCloudControllerServiceBrokerRepository(config, cloudControllerGateway)
	loc.servicePlanRepo = NewCloudControllerServicePlanRepository(config, cloudControllerGateway)
	loc.servicePlanVisibilityRepo = NewCloudControllerServicePlanVisibilityRepository(config, cloudControllerGateway)
	loc.serviceSummaryRepo = NewCloudControllerServiceSummaryRepository(config, cloudControllerGateway)
	loc.spaceRepo = spaces.NewCloudControllerSpaceRepository(config, cloudControllerGateway)
	loc.userProvidedServiceInstanceRepo = NewCCUserProvidedServiceInstanceRepository(config, cloudControllerGateway)
	loc.userRepo = NewCloudControllerUserRepository(config, uaaGateway, cloudControllerGateway)
	loc.clientRepo = NewCloudControllerClientRepository(config, uaaGateway)
	loc.buildpackRepo = NewCloudControllerBuildpackRepository(config, cloudControllerGateway)

	return
}

func (locator RepositoryLocator) SetAuthenticationRepository(repo authentication.Repository) RepositoryLocator {
	locator.authRepo = repo
	return locator
}

func (locator RepositoryLocator) GetAuthenticationRepository() authentication.Repository {
	return locator.authRepo
}

func (locator RepositoryLocator) SetCurlRepository(repo CurlRepository) RepositoryLocator {
	locator.curlRepo = repo
	return locator
}

func (locator RepositoryLocator) GetCurlRepository() CurlRepository {
	return locator.curlRepo
}

func (locator RepositoryLocator) GetEndpointRepository() coreconfig.EndpointRepository {
	return locator.endpointRepo
}

func (locator RepositoryLocator) SetEndpointRepository(e coreconfig.EndpointRepository) RepositoryLocator {
	locator.endpointRepo = e
	return locator
}

func (locator RepositoryLocator) SetOrganizationRepository(repo organizations.OrganizationRepository) RepositoryLocator {
	locator.organizationRepo = repo
	return locator
}

func (locator RepositoryLocator) GetOrganizationRepository() organizations.OrganizationRepository {
	return locator.organizationRepo
}

func (locator RepositoryLocator) SetSpaceRepository(repo spaces.SpaceRepository) RepositoryLocator {
	locator.spaceRepo = repo
	return locator
}

func (locator RepositoryLocator) GetSpaceRepository() spaces.SpaceRepository {
	return locator.spaceRepo
}

func (locator RepositoryLocator) SetApplicationRepository(repo applications.Repository) RepositoryLocator {
	locator.appRepo = repo
	return locator
}

func (locator RepositoryLocator) GetApplicationRepository() applications.Repository {
	return locator.appRepo
}

func (locator RepositoryLocator) SetAppSummaryRepository(repo AppSummaryRepository) RepositoryLocator {
	locator.appSummaryRepo = repo
	return locator
}

func (locator RepositoryLocator) SetUserRepository(repo UserRepository) RepositoryLocator {
	locator.userRepo = repo
	return locator
}

func (locator RepositoryLocator) GetAppSummaryRepository() AppSummaryRepository {
	return locator.appSummaryRepo
}

func (locator RepositoryLocator) SetAppInstancesRepository(repo appinstances.Repository) RepositoryLocator {
	locator.appInstancesRepo = repo
	return locator
}

func (locator RepositoryLocator) GetAppInstancesRepository() appinstances.Repository {
	return locator.appInstancesRepo
}

func (locator RepositoryLocator) SetDomainRepository(repo DomainRepository) RepositoryLocator {
	locator.domainRepo = repo
	return locator
}

func (locator RepositoryLocator) GetDomainRepository() DomainRepository {
	return locator.domainRepo
}

func (locator RepositoryLocator) SetRouteRepository(repo RouteRepository) RepositoryLocator {
	locator.routeRepo = repo
	return locator
}

func (locator RepositoryLocator) GetRoutingAPIRepository() RoutingAPIRepository {
	return locator.routingAPIRepo
}

func (locator RepositoryLocator) SetRoutingAPIRepository(repo RoutingAPIRepository) RepositoryLocator {
	locator.routingAPIRepo = repo
	return locator
}

func (locator RepositoryLocator) GetRouteRepository() RouteRepository {
	return locator.routeRepo
}

func (locator RepositoryLocator) SetServiceRepository(repo ServiceRepository) RepositoryLocator {
	locator.serviceRepo = repo
	return locator
}

func (locator RepositoryLocator) GetServiceRepository() ServiceRepository {
	return locator.serviceRepo
}

func (locator RepositoryLocator) SetServiceKeyRepository(repo ServiceKeyRepository) RepositoryLocator {
	locator.serviceKeyRepo = repo
	return locator
}

func (locator RepositoryLocator) GetServiceKeyRepository() ServiceKeyRepository {
	return locator.serviceKeyRepo
}

func (locator RepositoryLocator) SetRouteServiceBindingRepository(repo RouteServiceBindingRepository) RepositoryLocator {
	locator.routeServiceBindingRepo = repo
	return locator
}

func (locator RepositoryLocator) GetRouteServiceBindingRepository() RouteServiceBindingRepository {
	return locator.routeServiceBindingRepo
}

func (locator RepositoryLocator) SetServiceBindingRepository(repo ServiceBindingRepository) RepositoryLocator {
	locator.serviceBindingRepo = repo
	return locator
}

func (locator RepositoryLocator) GetServiceBindingRepository() ServiceBindingRepository {
	return locator.serviceBindingRepo
}

func (locator RepositoryLocator) GetServiceSummaryRepository() ServiceSummaryRepository {
	return locator.serviceSummaryRepo
}
func (locator RepositoryLocator) SetServiceSummaryRepository(repo ServiceSummaryRepository) RepositoryLocator {
	locator.serviceSummaryRepo = repo
	return locator
}

func (locator RepositoryLocator) GetUserRepository() UserRepository {
	return locator.userRepo
}

func (locator RepositoryLocator) GetClientRepository() ClientRepository {
	return locator.clientRepo
}

func (locator RepositoryLocator) SetClientRepository(repo ClientRepository) RepositoryLocator {
	locator.clientRepo = repo
	return locator
}

func (locator RepositoryLocator) SetPasswordRepository(repo password.Repository) RepositoryLocator {
	locator.passwordRepo = repo
	return locator
}

func (locator RepositoryLocator) GetPasswordRepository() password.Repository {
	return locator.passwordRepo
}

func (locator RepositoryLocator) SetServiceAuthTokenRepository(repo ServiceAuthTokenRepository) RepositoryLocator {
	locator.authTokenRepo = repo
	return locator
}

func (locator RepositoryLocator) GetServiceAuthTokenRepository() ServiceAuthTokenRepository {
	return locator.authTokenRepo
}

func (locator RepositoryLocator) SetServiceBrokerRepository(repo ServiceBrokerRepository) RepositoryLocator {
	locator.serviceBrokerRepo = repo
	return locator
}

func (locator RepositoryLocator) GetServiceBrokerRepository() ServiceBrokerRepository {
	return locator.serviceBrokerRepo
}

func (locator RepositoryLocator) GetServicePlanRepository() ServicePlanRepository {
	return locator.servicePlanRepo
}

func (locator RepositoryLocator) SetUserProvidedServiceInstanceRepository(repo UserProvidedServiceInstanceRepository) RepositoryLocator {
	locator.userProvidedServiceInstanceRepo = repo
	return locator
}

func (locator RepositoryLocator) GetUserProvidedServiceInstanceRepository() UserProvidedServiceInstanceRepository {
	return locator.userProvidedServiceInstanceRepo
}

func (locator RepositoryLocator) SetBuildpackRepository(repo BuildpackRepository) RepositoryLocator {
	locator.buildpackRepo = repo
	return locator
}

func (locator RepositoryLocator) GetBuildpackRepository() BuildpackRepository {
	return locator.buildpackRepo
}

func (locator RepositoryLocator) GetServicePlanVisibilityRepository() ServicePlanVisibilityRepository {
	return locator.servicePlanVisibilityRepo
}
