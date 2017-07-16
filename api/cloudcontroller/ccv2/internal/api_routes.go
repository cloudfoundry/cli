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
	DeleteOrganizationRequest              = "DeleteOrganization"
	DeleteRouteRequest                     = "DeleteRoute"
	DeleteRunningSecurityGroupSpaceRequest = "DeleteRunningSecurityGroupSpace"
	DeleteSecurityGroupSpaceRequest        = "DeleteSecurityGroupSpace"
	DeleteServiceBindingRequest            = "DeleteServiceBinding"
	DeleteSpaceRequest                     = "DeleteSpaceRequest"
	DeleteStagingSecurityGroupSpaceRequest = "DeleteStagingSecurityGroupSpace"
	GetAppInstancesRequest                 = "GetAppInstances"
	GetAppRequest                          = "GetApp"
	GetAppRoutesRequest                    = "GetAppRoutes"
	GetAppsRequest                         = "GetApps"
	GetAppStatsRequest                     = "GetAppStats"
	GetInfoRequest                         = "GetInfo"
	GetJobRequest                          = "GetJob"
	GetOrganizationPrivateDomainsRequest   = "GetOrganizationPrivateDomains"
	GetOrganizationQuotaDefinitionRequest  = "GetOrganizationQuotaDefinition"
	GetOrganizationRequest                 = "GetOrganization"
	GetOrganizationsRequest                = "GetOrganizations"
	GetPrivateDomainRequest                = "GetPrivateDomain"
	GetRouteAppsRequest                    = "GetRouteApps"
	GetRouteReservedRequest                = "GetRouteReserved"
	GetRouteRouteMappingsRequest           = "GetRouteRouteMappings"
	GetRoutesRequest                       = "GetRoutes"
	GetSecurityGroupRunningSpacesRequest   = "GetSecurityGroupRunningSpaces"
	GetSecurityGroupsRequest               = "GetSecurityGroups"
	GetSecurityGroupStagingSpacesRequest   = "GetSecurityGroupStagingSpaces"
	GetServiceBindingsRequest              = "GetServiceBindings"
	GetServiceInstancesRequest             = "GetServiceInstances"
	GetSharedDomainRequest                 = "GetSharedDomain"
	GetSharedDomainsRequest                = "GetSharedDomains"
	GetSpaceQuotaDefinitionRequest         = "GetSpaceQuotaDefinition"
	GetSpaceRoutesRequest                  = "GetSpaceRoutes"
	GetSpaceRunningSecurityGroupsRequest   = "GetSpaceRunningSecurityGroups"
	GetSpaceServiceInstancesRequest        = "GetSpaceServiceInstances"
	GetSpacesRequest                       = "GetSpaces"
	GetSpaceStagingSecurityGroupsRequest   = "GetSpaceStagingSecurityGroups"
	GetStackRequest                        = "GetStack"
	GetUsersRequest                        = "GetUsers"
	PostAppRequest                         = "PostApp"
	PostAppRestageRequest                  = "PostAppRestage"
	PostRouteRequest                       = "PostRoute"
	PostServiceBindingRequest              = "PostServiceBinding"
	PostUserRequest                        = "PostUser"
	PutAppBitsRequest                      = "PutAppBits"
	PutAppRequest                          = "PutApp"
	PutBindRouteAppRequest                 = "PutBindRouteApp"
	PutResourceMatch                       = "PutResourceMatch"
	PutRunningSecurityGroupSpaceRequest    = "PutRunningSecurityGroupSpace"
	PutStagingSecurityGroupSpaceRequest    = "PutStagingSecurityGroupSpace"
)

// APIRoutes is a list of routes used by the rata library to construct request
// URLs.
var APIRoutes = rata.Routes{
	{Path: "/v2/apps", Method: http.MethodGet, Name: GetAppsRequest},
	{Path: "/v2/apps", Method: http.MethodPost, Name: PostAppRequest},
	{Path: "/v2/apps/:app_guid", Method: http.MethodGet, Name: GetAppRequest},
	{Path: "/v2/apps/:app_guid", Method: http.MethodPut, Name: PutAppRequest},
	{Path: "/v2/apps/:app_guid/bits", Method: http.MethodPut, Name: PutAppBitsRequest},
	{Path: "/v2/apps/:app_guid/instances", Method: http.MethodGet, Name: GetAppInstancesRequest},
	{Path: "/v2/apps/:app_guid/routes", Method: http.MethodGet, Name: GetAppRoutesRequest},
	{Path: "/v2/apps/:app_guid/stats", Method: http.MethodGet, Name: GetAppStatsRequest},
	{Path: "/v2/apps/:app_guid/restage", Method: http.MethodPost, Name: PostAppRestageRequest},
	{Path: "/v2/info", Method: http.MethodGet, Name: GetInfoRequest},
	{Path: "/v2/jobs/:job_guid", Method: http.MethodGet, Name: GetJobRequest},
	{Path: "/v2/organizations", Method: http.MethodGet, Name: GetOrganizationsRequest},
	{Path: "/v2/organizations/:organization_guid", Method: http.MethodDelete, Name: DeleteOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid", Method: http.MethodGet, Name: GetOrganizationRequest},
	{Path: "/v2/organizations/:organization_guid/private_domains", Method: http.MethodGet, Name: GetOrganizationPrivateDomainsRequest},
	{Path: "/v2/private_domains/:private_domain_guid", Method: http.MethodGet, Name: GetPrivateDomainRequest},
	{Path: "/v2/quota_definitions/:organization_quota_guid", Method: http.MethodGet, Name: GetOrganizationQuotaDefinitionRequest},
	{Path: "/v2/resource_match", Method: http.MethodPut, Name: PutResourceMatch},
	{Path: "/v2/routes", Method: http.MethodGet, Name: GetRoutesRequest},
	{Path: "/v2/routes", Method: http.MethodPost, Name: PostRouteRequest},
	{Path: "/v2/routes/:route_guid", Method: http.MethodDelete, Name: DeleteRouteRequest},
	{Path: "/v2/routes/:route_guid/apps", Method: http.MethodGet, Name: GetRouteAppsRequest},
	{Path: "/v2/routes/:route_guid/apps/:app_guid", Method: http.MethodPut, Name: PutBindRouteAppRequest},
	{Path: "/v2/routes/:route_guid/route_mappings", Method: http.MethodGet, Name: GetRouteRouteMappingsRequest},
	{Path: "/v2/routes/reserved/domain/:domain_guid", Method: http.MethodGet, Name: GetRouteReservedRequest},
	{Path: "/v2/security_groups", Method: http.MethodGet, Name: GetSecurityGroupsRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces", Method: http.MethodGet, Name: GetSecurityGroupRunningSpacesRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteRunningSecurityGroupSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/spaces/:space_guid", Method: http.MethodPut, Name: PutRunningSecurityGroupSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces", Method: http.MethodGet, Name: GetSecurityGroupStagingSpacesRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces/:space_guid", Method: http.MethodDelete, Name: DeleteStagingSecurityGroupSpaceRequest},
	{Path: "/v2/security_groups/:security_group_guid/staging_spaces/:space_guid", Method: http.MethodPut, Name: PutStagingSecurityGroupSpaceRequest},
	{Path: "/v2/service_bindings", Method: http.MethodGet, Name: GetServiceBindingsRequest},
	{Path: "/v2/service_bindings", Method: http.MethodPost, Name: PostServiceBindingRequest},
	{Path: "/v2/service_bindings/:service_binding_guid", Method: http.MethodDelete, Name: DeleteServiceBindingRequest},
	{Path: "/v2/service_instances", Method: http.MethodGet, Name: GetServiceInstancesRequest},
	{Path: "/v2/shared_domains", Method: http.MethodGet, Name: GetSharedDomainsRequest},
	{Path: "/v2/shared_domains/:shared_domain_guid", Method: http.MethodGet, Name: GetSharedDomainRequest},
	{Path: "/v2/space_quota_definitions/:space_quota_guid", Method: http.MethodGet, Name: GetSpaceQuotaDefinitionRequest},
	{Path: "/v2/spaces", Method: http.MethodGet, Name: GetSpacesRequest},
	{Path: "/v2/spaces/:guid/service_instances", Method: http.MethodGet, Name: GetSpaceServiceInstancesRequest},
	{Path: "/v2/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSpaceRequest},
	{Path: "/v2/spaces/:space_guid/routes", Method: http.MethodGet, Name: GetSpaceRoutesRequest},
	{Path: "/v2/spaces/:space_guid/security_groups", Method: http.MethodGet, Name: GetSpaceRunningSecurityGroupsRequest},
	{Path: "/v2/spaces/:space_guid/staging_security_groups", Method: http.MethodGet, Name: GetSpaceStagingSecurityGroupsRequest},
	{Path: "/v2/stacks/:stack_guid", Method: http.MethodGet, Name: GetStackRequest},
	{Path: "/v2/users", Method: http.MethodPost, Name: PostUserRequest},
}
