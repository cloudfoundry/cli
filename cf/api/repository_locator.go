package api

import (
	"crypto/tls"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"net/http"

	"github.com/cloudfoundry/cli/cf/api/authentication"
	"github.com/cloudfoundry/cli/cf/api/feature_flag"
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/api/security_groups"
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running"
	"github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging"
	securitygroupspaces "github.com/cloudfoundry/cli/cf/api/security_groups/spaces"
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/api/strategy"

	"github.com/cloudfoundry/cli/cf/app_files"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
	"github.com/cloudfoundry/cli/cf/terminal"
	consumer "github.com/cloudfoundry/loggregator_consumer"
)

type RepositoryLocator struct {
	authRepo                        authentication.AuthenticationRepository
	curlRepo                        CurlRepository
	endpointRepo                    RemoteEndpointRepository
	organizationRepo                organizations.CloudControllerOrganizationRepository
	quotaRepo                       quotas.CloudControllerQuotaRepository
	spaceRepo                       spaces.CloudControllerSpaceRepository
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
	servicePlanRepo                 CloudControllerServicePlanRepository
	servicePlanVisibilityRepo       ServicePlanVisibilityRepository
	userProvidedServiceInstanceRepo CCUserProvidedServiceInstanceRepository
	buildpackRepo                   CloudControllerBuildpackRepository
	buildpackBitsRepo               CloudControllerBuildpackBitsRepository
	securityGroupRepo               security_groups.SecurityGroupRepo
	stagingSecurityGroupRepo        staging.StagingSecurityGroupsRepo
	runningSecurityGroupRepo        running.RunningSecurityGroupsRepo
	securityGroupSpaceBinder        securitygroupspaces.SecurityGroupSpaceBinder
	spaceQuotaRepo                  space_quotas.SpaceQuotaRepository
	featureFlagRepo                 feature_flag.FeatureFlagRepository
}

func NewRepositoryLocator(config configuration.ReadWriter, gatewaysByName map[string]net.Gateway) (loc RepositoryLocator) {
	strategy := strategy.NewEndpointStrategy(config.ApiVersion())

	authGateway := gatewaysByName["auth"]
	cloudControllerGateway := gatewaysByName["cloud-controller"]
	uaaGateway := gatewaysByName["uaa"]
	loc.authRepo = authentication.NewUAAAuthenticationRepository(authGateway, config)

	// ensure gateway refreshers are set before passing them by value to repositories
	cloudControllerGateway.SetTokenRefresher(loc.authRepo)
	uaaGateway.SetTokenRefresher(loc.authRepo)

	tlsConfig := net.NewTLSConfig([]tls.Certificate{}, config.IsSSLDisabled())
	loggregatorConsumer := consumer.New(config.LoggregatorEndpoint(), tlsConfig, http.ProxyFromEnvironment)
	loggregatorConsumer.SetDebugPrinter(terminal.DebugPrinter{})

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
	loc.organizationRepo = organizations.NewCloudControllerOrganizationRepository(config, cloudControllerGateway)
	loc.passwordRepo = NewCloudControllerPasswordRepository(config, uaaGateway)
	loc.quotaRepo = quotas.NewCloudControllerQuotaRepository(config, cloudControllerGateway)
	loc.routeRepo = NewCloudControllerRouteRepository(config, cloudControllerGateway)
	loc.stackRepo = NewCloudControllerStackRepository(config, cloudControllerGateway)
	loc.serviceRepo = NewCloudControllerServiceRepository(config, cloudControllerGateway)
	loc.serviceBindingRepo = NewCloudControllerServiceBindingRepository(config, cloudControllerGateway)
	loc.serviceBrokerRepo = NewCloudControllerServiceBrokerRepository(config, cloudControllerGateway)
	loc.servicePlanRepo = NewCloudControllerServicePlanRepository(config, cloudControllerGateway)
	loc.servicePlanVisibilityRepo = NewCloudControllerServicePlanVisibilityRepository(config, cloudControllerGateway)
	loc.serviceSummaryRepo = NewCloudControllerServiceSummaryRepository(config, cloudControllerGateway)
	loc.spaceRepo = spaces.NewCloudControllerSpaceRepository(config, cloudControllerGateway)
	loc.userProvidedServiceInstanceRepo = NewCCUserProvidedServiceInstanceRepository(config, cloudControllerGateway)
	loc.userRepo = NewCloudControllerUserRepository(config, uaaGateway, cloudControllerGateway)
	loc.buildpackRepo = NewCloudControllerBuildpackRepository(config, cloudControllerGateway)
	loc.buildpackBitsRepo = NewCloudControllerBuildpackBitsRepository(config, cloudControllerGateway, app_files.ApplicationZipper{})
	loc.securityGroupRepo = security_groups.NewSecurityGroupRepo(config, cloudControllerGateway)
	loc.stagingSecurityGroupRepo = staging.NewStagingSecurityGroupsRepo(config, cloudControllerGateway)
	loc.runningSecurityGroupRepo = running.NewRunningSecurityGroupsRepo(config, cloudControllerGateway)
	loc.securityGroupSpaceBinder = securitygroupspaces.NewSecurityGroupSpaceBinder(config, cloudControllerGateway)
	loc.spaceQuotaRepo = space_quotas.NewCloudControllerSpaceQuotaRepository(config, cloudControllerGateway)
	loc.featureFlagRepo = feature_flag.NewCloudControllerFeatureFlagRepository(config, cloudControllerGateway)
	return
}

func (locator RepositoryLocator) GetAuthenticationRepository() authentication.AuthenticationRepository {
	return locator.authRepo
}

func (locator RepositoryLocator) GetCurlRepository() CurlRepository {
	return locator.curlRepo
}

func (locator RepositoryLocator) GetEndpointRepository() EndpointRepository {
	return locator.endpointRepo
}

func (locator RepositoryLocator) GetOrganizationRepository() organizations.OrganizationRepository {
	return locator.organizationRepo
}

func (locator RepositoryLocator) GetQuotaRepository() quotas.QuotaRepository {
	return locator.quotaRepo
}

func (locator RepositoryLocator) GetSpaceRepository() spaces.SpaceRepository {
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

func (locator RepositoryLocator) GetServicePlanRepository() ServicePlanRepository {
	return locator.servicePlanRepo
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

func (locator RepositoryLocator) GetSecurityGroupRepository() security_groups.SecurityGroupRepo {
	return locator.securityGroupRepo
}

func (locator RepositoryLocator) GetStagingSecurityGroupsRepository() staging.StagingSecurityGroupsRepo {
	return locator.stagingSecurityGroupRepo
}

func (locator RepositoryLocator) GetRunningSecurityGroupsRepository() running.RunningSecurityGroupsRepo {
	return locator.runningSecurityGroupRepo
}

func (locator RepositoryLocator) GetSecurityGroupSpaceBinder() securitygroupspaces.SecurityGroupSpaceBinder {
	return locator.securityGroupSpaceBinder
}

func (locator RepositoryLocator) GetServicePlanVisibilityRepository() ServicePlanVisibilityRepository {
	return locator.servicePlanVisibilityRepo
}

func (locator RepositoryLocator) GetSpaceQuotaRepository() space_quotas.SpaceQuotaRepository {
	return locator.spaceQuotaRepo
}

func (locator RepositoryLocator) GetFeatureFlagRepository() feature_flag.FeatureFlagRepository {
	return locator.featureFlagRepo
}
