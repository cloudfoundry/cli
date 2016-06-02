package api

import (
	"crypto/tls"
	"net/http"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/appevents"
	api_appfiles "github.com/cloudfoundry/cli/cf/api/appfiles"
	"github.com/cloudfoundry/cli/cf/api/appinstances"
	"github.com/cloudfoundry/cli/cf/api/applicationbits"
	"github.com/cloudfoundry/cli/cf/api/applications"
	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/copyapplicationsource"
	"github.com/cloudfoundry/cli/cf/api/environmentvariablegroups"
	"github.com/cloudfoundry/cli/cf/api/featureflags"
	"github.com/cloudfoundry/cli/cf/api/logs"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/password"
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/api/securitygroups"
	"github.com/cloudfoundry/cli/cf/api/securitygroups/defaults/running"
	"github.com/cloudfoundry/cli/cf/api/securitygroups/defaults/staging"
	securitygroupspaces "github.com/cloudfoundry/cli/cf/api/securitygroups/spaces"
	"github.com/cloudfoundry/cli/cf/api/spacequotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/api/stacks"
	"github.com/cloudfoundry/cli/cf/api/strategy"
	"github.com/cloudfoundry/cli/cf/appfiles"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/cf/trace"
	"github.com/cloudfoundry/cli/cf/v3/repository"
	v3client "github.com/cloudfoundry/go-ccapi/v3/client"
	"github.com/cloudfoundry/loggregator_consumer"
	"github.com/cloudfoundry/noaa/consumer"
)

type RepositoryLocator struct {
	authRepo                        authentication.Repository
	curlRepo                        CurlRepository
	endpointRepo                    coreconfig.EndpointRepository
	organizationRepo                organizations.OrganizationRepository
	quotaRepo                       quotas.QuotaRepository
	spaceRepo                       spaces.SpaceRepository
	appRepo                         applications.Repository
	appBitsRepo                     applicationbits.CloudControllerApplicationBitsRepository
	appSummaryRepo                  AppSummaryRepository
	appInstancesRepo                appinstances.Repository
	appEventsRepo                   appevents.Repository
	appFilesRepo                    api_appfiles.Repository
	domainRepo                      DomainRepository
	routeRepo                       RouteRepository
	routingAPIRepo                  RoutingAPIRepository
	stackRepo                       stacks.StackRepository
	serviceRepo                     ServiceRepository
	serviceKeyRepo                  ServiceKeyRepository
	serviceBindingRepo              ServiceBindingRepository
	routeServiceBindingRepo         RouteServiceBindingRepository
	serviceSummaryRepo              ServiceSummaryRepository
	userRepo                        UserRepository
	passwordRepo                    password.Repository
	logsRepo                        logs.Repository
	authTokenRepo                   ServiceAuthTokenRepository
	serviceBrokerRepo               ServiceBrokerRepository
	servicePlanRepo                 CloudControllerServicePlanRepository
	servicePlanVisibilityRepo       ServicePlanVisibilityRepository
	userProvidedServiceInstanceRepo UserProvidedServiceInstanceRepository
	buildpackRepo                   BuildpackRepository
	buildpackBitsRepo               BuildpackBitsRepository
	securityGroupRepo               securitygroups.SecurityGroupRepo
	stagingSecurityGroupRepo        staging.SecurityGroupsRepo
	runningSecurityGroupRepo        running.SecurityGroupsRepo
	securityGroupSpaceBinder        securitygroupspaces.SecurityGroupSpaceBinder
	spaceQuotaRepo                  spacequotas.SpaceQuotaRepository
	featureFlagRepo                 featureflags.FeatureFlagRepository
	environmentVariableGroupRepo    environmentvariablegroups.Repository
	copyAppSourceRepo               copyapplicationsource.Repository

	v3Repository repository.Repository
}

func NewRepositoryLocator(config coreconfig.ReadWriter, gatewaysByName map[string]net.Gateway, logger trace.Printer) (loc RepositoryLocator) {
	strategy := strategy.NewEndpointStrategy(config.APIVersion())

	cloudControllerGateway := gatewaysByName["cloud-controller"]
	routingAPIGateway := gatewaysByName["routing-api"]
	uaaGateway := gatewaysByName["uaa"]
	loc.authRepo = authentication.NewUAARepository(uaaGateway, config, net.NewRequestDumper(logger))

	// ensure gateway refreshers are set before passing them by value to repositories
	cloudControllerGateway.SetTokenRefresher(loc.authRepo)
	uaaGateway.SetTokenRefresher(loc.authRepo)

	loc.appBitsRepo = applicationbits.NewCloudControllerApplicationBitsRepository(config, cloudControllerGateway)
	loc.appEventsRepo = appevents.NewCloudControllerAppEventsRepository(config, cloudControllerGateway, strategy)
	loc.appFilesRepo = api_appfiles.NewCloudControllerAppFilesRepository(config, cloudControllerGateway)
	loc.appRepo = applications.NewCloudControllerRepository(config, cloudControllerGateway)
	loc.appSummaryRepo = NewCloudControllerAppSummaryRepository(config, cloudControllerGateway)
	loc.appInstancesRepo = appinstances.NewCloudControllerAppInstancesRepository(config, cloudControllerGateway)
	loc.authTokenRepo = NewCloudControllerServiceAuthTokenRepository(config, cloudControllerGateway)
	loc.curlRepo = NewCloudControllerCurlRepository(config, cloudControllerGateway)
	loc.domainRepo = NewCloudControllerDomainRepository(config, cloudControllerGateway, strategy)
	loc.endpointRepo = NewEndpointRepository(cloudControllerGateway)

	tlsConfig := net.NewTLSConfig([]tls.Certificate{}, config.IsSSLDisabled())

	apiVersion, _ := semver.Make(config.APIVersion())

	if apiVersion.GTE(cf.NoaaMinimumAPIVersion) {
		consumer := consumer.New(config.DopplerEndpoint(), tlsConfig, http.ProxyFromEnvironment)
		consumer.SetDebugPrinter(terminal.DebugPrinter{Logger: logger})
		loc.logsRepo = logs.NewNoaaLogsRepository(config, consumer, loc.authRepo)
	} else {
		consumer := loggregator_consumer.New(config.LoggregatorEndpoint(), tlsConfig, http.ProxyFromEnvironment)
		consumer.SetDebugPrinter(terminal.DebugPrinter{Logger: logger})
		loc.logsRepo = logs.NewLoggregatorLogsRepository(config, consumer, loc.authRepo)
	}

	loc.organizationRepo = organizations.NewCloudControllerOrganizationRepository(config, cloudControllerGateway)
	loc.passwordRepo = password.NewCloudControllerRepository(config, uaaGateway)
	loc.quotaRepo = quotas.NewCloudControllerQuotaRepository(config, cloudControllerGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, cloudControllerGateway)
	loc.routeServiceBindingRepo = NewCloudControllerRouteServiceBindingRepository(config, cloudControllerGateway)
	loc.routingAPIRepo = NewRoutingAPIRepository(config, routingAPIGateway)
	loc.stackRepo = stacks.NewCloudControllerStackRepository(config, cloudControllerGateway)
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
	loc.buildpackRepo = NewCloudControllerBuildpackRepository(config, cloudControllerGateway)
	loc.buildpackBitsRepo = NewCloudControllerBuildpackBitsRepository(config, cloudControllerGateway, appfiles.ApplicationZipper{})
	loc.securityGroupRepo = securitygroups.NewSecurityGroupRepo(config, cloudControllerGateway)
	loc.stagingSecurityGroupRepo = staging.NewSecurityGroupsRepo(config, cloudControllerGateway)
	loc.runningSecurityGroupRepo = running.NewSecurityGroupsRepo(config, cloudControllerGateway)
	loc.securityGroupSpaceBinder = securitygroupspaces.NewSecurityGroupSpaceBinder(config, cloudControllerGateway)
	loc.spaceQuotaRepo = spacequotas.NewCloudControllerSpaceQuotaRepository(config, cloudControllerGateway)
	loc.featureFlagRepo = featureflags.NewCloudControllerFeatureFlagRepository(config, cloudControllerGateway)
	loc.environmentVariableGroupRepo = environmentvariablegroups.NewCloudControllerRepository(config, cloudControllerGateway)
	loc.copyAppSourceRepo = copyapplicationsource.NewCloudControllerCopyApplicationSourceRepository(config, cloudControllerGateway)

	client := v3client.NewClient(config.APIEndpoint(), config.AuthenticationEndpoint(), config.AccessToken(), config.RefreshToken())
	loc.v3Repository = repository.NewRepository(config, client)

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

func (locator RepositoryLocator) SetQuotaRepository(repo quotas.QuotaRepository) RepositoryLocator {
	locator.quotaRepo = repo
	return locator
}

func (locator RepositoryLocator) GetQuotaRepository() quotas.QuotaRepository {
	return locator.quotaRepo
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

func (locator RepositoryLocator) GetApplicationBitsRepository() applicationbits.Repository {
	return locator.appBitsRepo
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

func (locator RepositoryLocator) SetAppEventsRepository(repo appevents.Repository) RepositoryLocator {
	locator.appEventsRepo = repo
	return locator
}

func (locator RepositoryLocator) GetAppEventsRepository() appevents.Repository {
	return locator.appEventsRepo
}

func (locator RepositoryLocator) SetAppFileRepository(repo api_appfiles.Repository) RepositoryLocator {
	locator.appFilesRepo = repo
	return locator
}

func (locator RepositoryLocator) GetAppFilesRepository() api_appfiles.Repository {
	return locator.appFilesRepo
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

func (locator RepositoryLocator) SetStackRepository(repo stacks.StackRepository) RepositoryLocator {
	locator.stackRepo = repo
	return locator
}

func (locator RepositoryLocator) GetStackRepository() stacks.StackRepository {
	return locator.stackRepo
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

func (locator RepositoryLocator) SetPasswordRepository(repo password.Repository) RepositoryLocator {
	locator.passwordRepo = repo
	return locator
}

func (locator RepositoryLocator) GetPasswordRepository() password.Repository {
	return locator.passwordRepo
}

func (locator RepositoryLocator) SetLogsRepository(repo logs.Repository) RepositoryLocator {
	locator.logsRepo = repo
	return locator
}

func (locator RepositoryLocator) GetLogsRepository() logs.Repository {
	return locator.logsRepo
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

func (locator RepositoryLocator) SetBuildpackBitsRepository(repo BuildpackBitsRepository) RepositoryLocator {
	locator.buildpackBitsRepo = repo
	return locator
}

func (locator RepositoryLocator) GetBuildpackBitsRepository() BuildpackBitsRepository {
	return locator.buildpackBitsRepo
}

func (locator RepositoryLocator) SetSecurityGroupRepository(repo securitygroups.SecurityGroupRepo) RepositoryLocator {
	locator.securityGroupRepo = repo
	return locator
}

func (locator RepositoryLocator) GetSecurityGroupRepository() securitygroups.SecurityGroupRepo {
	return locator.securityGroupRepo
}

func (locator RepositoryLocator) SetStagingSecurityGroupRepository(repo staging.SecurityGroupsRepo) RepositoryLocator {
	locator.stagingSecurityGroupRepo = repo
	return locator
}

func (locator RepositoryLocator) GetStagingSecurityGroupsRepository() staging.SecurityGroupsRepo {
	return locator.stagingSecurityGroupRepo
}

func (locator RepositoryLocator) SetRunningSecurityGroupRepository(repo running.SecurityGroupsRepo) RepositoryLocator {
	locator.runningSecurityGroupRepo = repo
	return locator
}

func (locator RepositoryLocator) GetRunningSecurityGroupsRepository() running.SecurityGroupsRepo {
	return locator.runningSecurityGroupRepo
}

func (locator RepositoryLocator) SetSecurityGroupSpaceBinder(repo securitygroupspaces.SecurityGroupSpaceBinder) RepositoryLocator {
	locator.securityGroupSpaceBinder = repo
	return locator
}

func (locator RepositoryLocator) GetSecurityGroupSpaceBinder() securitygroupspaces.SecurityGroupSpaceBinder {
	return locator.securityGroupSpaceBinder
}

func (locator RepositoryLocator) GetServicePlanVisibilityRepository() ServicePlanVisibilityRepository {
	return locator.servicePlanVisibilityRepo
}

func (locator RepositoryLocator) GetSpaceQuotaRepository() spacequotas.SpaceQuotaRepository {
	return locator.spaceQuotaRepo
}

func (locator RepositoryLocator) SetSpaceQuotaRepository(repo spacequotas.SpaceQuotaRepository) RepositoryLocator {
	locator.spaceQuotaRepo = repo
	return locator
}

func (locator RepositoryLocator) SetFeatureFlagRepository(repo featureflags.FeatureFlagRepository) RepositoryLocator {
	locator.featureFlagRepo = repo
	return locator
}

func (locator RepositoryLocator) GetFeatureFlagRepository() featureflags.FeatureFlagRepository {
	return locator.featureFlagRepo
}

func (locator RepositoryLocator) SetEnvironmentVariableGroupsRepository(repo environmentvariablegroups.Repository) RepositoryLocator {
	locator.environmentVariableGroupRepo = repo
	return locator
}

func (locator RepositoryLocator) GetEnvironmentVariableGroupsRepository() environmentvariablegroups.Repository {
	return locator.environmentVariableGroupRepo
}

func (locator RepositoryLocator) SetCopyApplicationSourceRepository(repo copyapplicationsource.Repository) RepositoryLocator {
	locator.copyAppSourceRepo = repo
	return locator
}

func (locator RepositoryLocator) GetCopyApplicationSourceRepository() copyapplicationsource.Repository {
	return locator.copyAppSourceRepo
}

func (locator RepositoryLocator) GetV3Repository() repository.Repository {
	return locator.v3Repository
}

func (locator RepositoryLocator) SetV3Repository(r repository.Repository) RepositoryLocator {
	locator.v3Repository = r
	return locator
}
