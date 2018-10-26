package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

// Naming convention:
//
// Method + non-parameter parts of the path
//
// If the request returns a single entity by GUID, use the singular (for example
// /v2/organizations/:organization_guid is GetOrganization).
//
// The const name should always be the const value + Request.
const (
	DeleteOrganizationRequest                            = "DeleteOrganization"
	DeleteRouteAppRequest                                = "DeleteRouteApp"
	DeleteRouteRequest                                   = "DeleteRoute"
	DeleteSecurityGroupSpaceRequest                      = "DeleteSecurityGroupSpace"
	DeleteSecurityGroupStagingSpaceRequest               = "DeleteSecurityGroupStagingSpace"
	DeleteServiceBindingRequest                          = "DeleteServiceBinding"
	DeleteSpaceRequest                                   = "DeleteSpace"
	GetAppInstancesRequest                               = "GetAppInstances"
	GetAppRequest                                        = "GetApp"
	GetAppRoutesRequest                                  = "GetAppRoutes"
	GetAppsRequest                                       = "GetApps"
	GetAppStatsRequest                                   = "GetAppStats"
	GetBuildpacksRequest                                 = "GetBuildpacks"
	GetConfigFeatureFlagsRequest                         = "GetConfigFeatureFlags"
	GetEventsRequest                                     = "GetEvents"
	GetInfoRequest                                       = "GetInfo"
	GetJobRequest                                        = "GetJob"
	GetOrganizationPrivateDomainsRequest                 = "GetOrganizationPrivateDomains"
	GetOrganizationQuotaDefinitionsRequest               = "GetOrganizationQuotaDefinitions"
	GetOrganizationQuotaDefinitionRequest                = "GetOrganizationQuotaDefinition"
	GetOrganizationRequest                               = "GetOrganization"
	GetOrganizationsRequest                              = "GetOrganizations"
	GetPrivateDomainRequest                              = "GetPrivateDomain"
	GetPrivateDomainsRequest                             = "GetPrivateDomains"
	GetRouteAppsRequest                                  = "GetRouteApps"
	GetRouteMappingRequest                               = "GetRouteMapping"
	GetRouteMappingsRequest                              = "GetRouteMappings"
	GetRouteRequest                                      = "GetRoute"
	GetRouteReservedDeprecatedRequest                    = "GetRouteReservedDeprecated"
	GetRouteReservedRequest                              = "GetRouteReserved"
	GetRouteRouteMappingsRequest                         = "GetRouteRouteMappings"
	GetRoutesRequest                                     = "GetRoutes"
	GetSecurityGroupSpacesRequest                        = "GetSecurityGroupSpaces"
	GetSecurityGroupsRequest                             = "GetSecurityGroups"
	GetSecurityGroupStagingSpacesRequest                 = "GetSecurityGroupStagingSpaces"
	GetServiceBindingRequest                             = "GetServiceBinding"
	GetServiceBindingsRequest                            = "GetServiceBindings"
	GetServiceBrokersRequest                             = "GetServiceBrokers"
	GetServiceInstanceRequest                            = "GetServiceInstance"
	GetServiceInstanceServiceBindingsRequest             = "GetServiceInstanceServiceBindings"
	GetServiceInstanceSharedFromRequest                  = "GetServiceInstanceSharedFrom"
	GetServiceInstanceSharedToRequest                    = "GetServiceInstanceSharedTo"
	GetServiceInstancesRequest                           = "GetServiceInstances"
	GetServicePlanRequest                                = "GetServicePlan"
	GetServicePlansRequest                               = "GetServicePlans"
	GetServicePlanVisibilitiesRequest                    = "GetServicePlanVisibilities"
	GetServiceRequest                                    = "GetService"
	GetServicesRequest                                   = "GetServices"
	GetSharedDomainRequest                               = "GetSharedDomain"
	GetSharedDomainsRequest                              = "GetSharedDomains"
	GetOrganizationSpaceQuotasRequest                    = "GetOrganizationSpaceQuotas"
	GetSpaceQuotaDefinitionRequest                       = "GetSpaceQuotaDefinition"
	GetSpaceRoutesRequest                                = "GetSpaceRoutes"
	GetSpaceSecurityGroupsRequest                        = "GetSpaceSecurityGroups"
	GetSpaceServiceInstancesRequest                      = "GetSpaceServiceInstances"
	GetSpacesRequest                                     = "GetSpaces"
	GetSpaceStagingSecurityGroupsRequest                 = "GetSpaceStagingSecurityGroups"
	GetStackRequest                                      = "GetStack"
	GetStacksRequest                                     = "GetStacks"
	GetUserProvidedServiceInstanceServiceBindingsRequest = "GetUserProvidedServiceInstanceServiceBindings"
	GetUserProvidedServiceInstancesRequest               = "GetUserProvidedServiceInstances"
	GetUsersRequest                                      = "GetUsers"
	PostAppRequest                                       = "PostApp"
	PostAppRestageRequest                                = "PostAppRestage"
	PostBuildpackRequest                                 = "PostBuildpack"
	PostOrganizationRequest                              = "PostOrganization"
	PostRouteRequest                                     = "PostRoute"
	PostServiceBindingRequest                            = "PostServiceBinding"
	PostSharedDomainRequest                              = "PostSharedDomain"
	PostServiceKeyRequest                                = "PostServiceKey"
	PostSpaceRequest                                     = "PostSpace"
	PostUserRequest                                      = "PostUser"
	PutAppBitsRequest                                    = "PutAppBits"
	PutAppRequest                                        = "PutApp"
	PutBuildpackRequest                                  = "PutBuildpack"
	PutBuildpackBitsRequest                              = "PutBuildpackBits"
	PutDropletRequest                                    = "PutDroplet"
	PutOrganizationManagerByUsernameRequest              = "PutOrganizationManagerByUsername"
	PutOrganizationManagerRequest                        = "PutOrganizationManager"
	PutOrganizationUserRequest                           = "PutOrganizationUser"
	PutOrganizationUserByUsernameRequest                 = "PutOrganizationUserByUsername"
	PutResourceMatchRequest                              = "PutResourceMatch"
	PutRouteAppRequest                                   = "PutRouteApp"
	PutSpaceQuotaRequest                                 = "PutSpaceQuotaRequest"
	PutSpaceDeveloperRequest                             = "PutSpaceDeveloper"
	PutSpaceDeveloperByUsernameRequest                   = "PutSpaceDeveloperByUsername"
	PutSpaceManagerRequest                               = "PutSpaceManager"
	PutSpaceManagerByUsernameRequest                     = "PutSpaceManagerByUsername"
	PutSecurityGroupSpaceRequest                         = "PutSecurityGroupSpace"
	PutSecurityGroupStagingSpaceRequest                  = "PutSecurityGroupStagingSpace"
)

// APIRoutes is a list of routes used by the rata library to construct request
// URLs.
var APIRoutes = rata.Routes{
	{Path: "/v2/apps", Method: http.MethodGet, Name: GetAppsRequest},
	{Path: "/v2/apps", Method: http.MethodPost, Name: PostAppRequest},
	{Path: "/v2/apps/:app_guid", Method: http.MethodGet, Name: GetAppRequest},
	{Path: "/v2/apps/:app_guid", Method: http.MethodPut, Name: PutAppRequest},
	{Path: "/v2/apps/:app_guid/bits", Method: http.MethodPut, Name: PutAppBitsRequest},
	{Path: "/v2/apps/:app_guid/droplet/upload", Method: http.MethodPut, Name: PutDropletRequest},
	{Path: "/v2/apps/:app_guid/instances", Method: http.MethodGet, Name: GetAppInstancesRequest},
	{Path: "/v2/apps/:app_guid/restage", Method: http.MethodPost, Name: PostAppRestageRequest},
	{Path: "/v2/apps/:app_guid/routes", Method: http.MethodGet, Name: GetAppRoutesRequest},
	{Path: "/v2/apps/:app_guid/stats", Method: http.MethodGet, Name: GetAppStatsRequest},
	{Path: "/v2/buildpacks", Method: http.MethodPost, Name: PostBuildpackRequest},
	{Path: "/v2/buildpacks", Method: http.MethodGet, Name: GetBuildpacksRequest},
	{Path: "/v2/buildpacks/:buildpack_guid", Method: http.MethodPut, Name: PutBuildpackRequest},
	{Path: "/v2/buildpacks/:buildpack_guid/bits", Method: http.MethodPut, Name: PutBuildpackBitsRequest},
	{Path: "/v2/config/feature_flags", Method: http.MethodGet, Name: GetConfigFeatureFlagsRequest},
	{Path: "/v2/events", Method: http.MethodGet, Name: GetEventsRequest},
	{Path: "/v2/info", Method: http.MethodGet, Name: GetInfoRequest},
	{Path: "/v2/jobs/:job_guid", Method: http.MethodGet, Name: GetJobRequest},
	{Path: "/v2/organizations", Method: http.MethodGet, Name: GetOrganizationsRequest},
	{Path: "/v2/organizations", Method: http.MethodPost, Name: PostOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid", Method: http.MethodDelete, Name: DeleteOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid", Method: http.MethodGet, Name: GetOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid/managers", Method: http.MethodPut, Name: PutOrganizationManagerByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/managers/:manager_guid", Method: http.MethodPut, Name: PutOrganizationManagerRequest},
	{Path: "/v2/organizations/:organization_guid/private_domains", Method: http.MethodGet, Name: GetOrganizationPrivateDomainsRequest},
	{Path: "/v2/organizations/:organization_guid/users", Method: http.MethodPut, Name: PutOrganizationUserByUsernameRequest},
	{Path: "/v2/organizations/:organization_guid/users/:user_guid", Method: http.MethodPut, Name: PutOrganizationUserRequest},
	{Path: "/v2/private_domains", Method: http.MethodGet, Name: GetPrivateDomainsRequest},
	{Path: "/v2/private_domains/:private_domain_guid", Method: http.MethodGet, Name: GetPrivateDomainRequest},
	{Path: "/v2/quota_definitions/:organization_quota_guid", Method: http.MethodGet, Name: GetOrganizationQuotaDefinitionRequest},
	{Path: "/v2/quota_definitions", Method: http.MethodGet, Name: GetOrganizationQuotaDefinitionsRequest},
	{Path: "/v2/resource_match", Method: http.MethodPut, Name: PutResourceMatchRequest},
	{Path: "/v2/route_mappings", Method: http.MethodGet, Name: GetRouteMappingsRequest},
	{Path: "/v2/route_mappings/:route_mapping_guid", Method: http.MethodGet, Name: GetRouteMappingRequest},
	{Path: "/v2/routes", Method: http.MethodGet, Name: GetRoutesRequest},
	{Path: "/v2/routes", Method: http.MethodPost, Name: PostRouteRequest},
	{Path: "/v2/routes/:route_guid", Method: http.MethodDelete, Name: DeleteRouteRequest},
	{Path: "/v2/routes/:route_guid", Method: http.MethodGet, Name: GetRouteRequest},
	{Path: "/v2/routes/:route_guid/apps", Method: http.MethodGet, Name: GetRouteAppsRequest},
	{Path: "/v2/routes/:route_guid/apps/:app_guid", Method: http.MethodDelete, Name: DeleteRouteAppRequest},
	{Path: "/v2/routes/:route_guid/apps/:app_guid", Method: http.MethodPut, Name: PutRouteAppRequest},
	{Path: "/v2/routes/:route_guid/route_mappings", Method: http.MethodGet, Name: GetRouteRouteMappingsRequest},
	{Path: "/v2/routes/reserved/domain/:domain_guid", Method: http.MethodGet, Name: GetRouteReservedRequest},
	{Path: "/v2/routes/reserved/domain/:domain_guid/host/:host", Method: http.MethodGet, Name: GetRouteReservedDeprecatedRequest},
	{Path: "/v2/security_groups", Method: http.MethodGet, Name: GetSecurityGroupsRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces", Method: http.MethodGet, Name: GetSecurityGroupSpacesRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSecurityGroupSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces/:space_guid", Method: http.MethodPut, Name: PutSecurityGroupSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces", Method: http.MethodGet, Name: GetSecurityGroupStagingSpacesRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSecurityGroupStagingSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces/:space_guid", Method: http.MethodPut, Name: PutSecurityGroupStagingSpaceRequest},
	{Path: "/v2/service_bindings", Method: http.MethodGet, Name: GetServiceBindingsRequest},
	{Path: "/v2/service_bindings", Method: http.MethodPost, Name: PostServiceBindingRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: http.MethodDelete, Name: DeleteServiceBindingRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: http.MethodGet, Name: GetServiceBindingRequest},
	{Path: "/v2/service_brokers", Method: http.MethodGet, Name: GetServiceBrokersRequest},
	{Path: "/v2/service_instances", Method: http.MethodGet, Name: GetServiceInstancesRequest},
	{Path: "/v2/service_instances/:service_instance_guid", Method: http.MethodGet, Name: GetServiceInstanceRequest},
	{Path: "/v2/service_instances/:service_instance_guid/service_bindings", Method: http.MethodGet, Name: GetServiceInstanceServiceBindingsRequest},
	{Path: "/v2/service_instances/:service_instance_guid/shared_from", Method: http.MethodGet, Name: GetServiceInstanceSharedFromRequest},
	{Path: "/v2/service_instances/:service_instance_guid/shared_to", Method: http.MethodGet, Name: GetServiceInstanceSharedToRequest},
	{Path: "/v2/service_keys", Method: http.MethodPost, Name: PostServiceKeyRequest},
	{Path: "/v2/service_plan_visibilities", Method: http.MethodGet, Name: GetServicePlanVisibilitiesRequest},
	{Path: "/v2/service_plans", Method: http.MethodGet, Name: GetServicePlansRequest},
	{Path: "/v2/service_plans/:service_plan_guid", Method: http.MethodGet, Name: GetServicePlanRequest},
	{Path: "/v2/services", Method: http.MethodGet, Name: GetServicesRequest},
	{Path: "/v2/services/:service_guid", Method: http.MethodGet, Name: GetServiceRequest},
	{Path: "/v2/shared_domains", Method: http.MethodGet, Name: GetSharedDomainsRequest},
	{Path: "/v2/shared_domains", Method: http.MethodPost, Name: PostSharedDomainRequest},
	{Path: "/v2/shared_domains/:shared_domain_guid", Method: http.MethodGet, Name: GetSharedDomainRequest},
	{Path: "/v2/organizations/:organization_guid/space_quota_definitions", Method: http.MethodGet, Name: GetOrganizationSpaceQuotasRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid/spaces/:space_guid", Method: http.MethodPut, Name: PutSpaceQuotaRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid", Method: http.MethodGet, Name: GetSpaceQuotaDefinitionRequest},
	{Path: "/v2/spaces", Method: http.MethodGet, Name: GetSpacesRequest},
	{Path: "/v2/spaces", Method: http.MethodPost, Name: PostSpaceRequest},
	{Path: "/v2/spaces/:space_guid/developers", Method: http.MethodPut, Name: PutSpaceDeveloperByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/developers/:developer_guid", Method: http.MethodPut, Name: PutSpaceDeveloperRequest},
	{Path: "/v2/spaces/:guid/service_instances", Method: http.MethodGet, Name: GetSpaceServiceInstancesRequest},
	{Path: "/v2/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSpaceRequest},
	{Path: "/v2/spaces/:space_guid/routes", Method: http.MethodGet, Name: GetSpaceRoutesRequest},
	{Path: "/v2/spaces/:space_guid/security_groups", Method: http.MethodGet, Name: GetSpaceSecurityGroupsRequest},
	{Path: "/v2/spaces/:space_guid/staging_security_groups", Method: http.MethodGet, Name: GetSpaceStagingSecurityGroupsRequest},
	{Path: "/v2/spaces/:space_guid/managers", Method: http.MethodPut, Name: PutSpaceManagerByUsernameRequest},
	{Path: "/v2/spaces/:space_guid/managers/:manager_guid", Method: http.MethodPut, Name: PutSpaceManagerRequest},
	{Path: "/v2/stacks", Method: http.MethodGet, Name: GetStacksRequest},
	{Path: "/v2/stacks/:stack_guid", Method: http.MethodGet, Name: GetStackRequest},
	{Path: "/v2/user_provided_service_instances", Method: http.MethodGet, Name: GetUserProvidedServiceInstancesRequest},
	{Path: "/v2/user_provided_service_instances/:user_provided_service_instance_guid/service_bindings", Method: http.MethodGet, Name: GetUserProvidedServiceInstanceServiceBindingsRequest},
	{Path: "/v2/users", Method: http.MethodPost, Name: PostUserRequest},
}
