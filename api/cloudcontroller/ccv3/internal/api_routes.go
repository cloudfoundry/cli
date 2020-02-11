package internal

import "net/http"

// Naming convention:
//
// HTTP method + non-parameter parts of the path + "Request"
//
// If the request returns a single entity by GUID, use the singular (for example
// /v2/organizations/:organization_guid is GetOrganization).
const (
	DeleteApplicationProcessInstanceRequest                     = "DeleteApplicationProcessInstance"
	DeleteApplicationRequest                                    = "DeleteApplication"
	DeleteBuildpackRequest                                      = "DeleteBuildpack"
	DeleteDomainRequest                                         = "DeleteDomainRequest"
	DeleteIsolationSegmentRelationshipOrganizationRequest       = "DeleteIsolationSegmentRelationshipOrganization"
	DeleteIsolationSegmentRequest                               = "DeleteIsolationSegment"
	DeleteOrganizationRequest                                   = "DeleteOrganization"
	DeleteOrganizationQuotaRequest                              = "DeleteOrganizationQuota"
	DeleteOrphanedRoutesRequest                                 = "DeleteOrphanedRoutes"
	DeleteRouteRequest                                          = "DeleteRouteRequest"
	DeleteRoleRequest                                           = "DeleteRoleRequest"
	DeleteServiceBrokerRequest                                  = "DeleteServiceBrokerRequest"
	DeleteServiceInstanceRelationshipsSharedSpaceRequest        = "DeleteServiceInstanceRelationshipsSharedSpace"
	DeleteSharedOrgFromDomainRequest                            = "DeleteSharedOrgFromDomain"
	DeleteSpaceQuotaRequest                                     = "DeleteSpaceQuota"
	DeleteSpaceRequest                                          = "DeleteSpace"
	DeleteSpaceQuotaFromSpaceRequest                            = "DeleteSpaceQuotaFromSpace"
	DeleteUserRequest                                           = "DeleteUser"
	GetApplicationDropletCurrentRequest                         = "GetApplicationDropletCurrent"
	GetApplicationEnvRequest                                    = "GetApplicationEnv"
	GetApplicationFeaturesRequest                               = "GetApplicationFeatures"
	GetApplicationManifestRequest                               = "GetApplicationManifest"
	GetApplicationProcessRequest                                = "GetApplicationProcess"
	GetApplicationProcessesRequest                              = "GetApplicationProcesses"
	GetApplicationRoutesRequest                                 = "GetApplicationRoutes"
	GetApplicationTasksRequest                                  = "GetApplicationTasks"
	GetApplicationsRequest                                      = "GetApplications"
	GetBuildRequest                                             = "GetBuild"
	GetBuildpacksRequest                                        = "GetBuildpacks"
	GetDefaultDomainRequest                                     = "GetDefaultDomain"
	GetDeploymentRequest                                        = "GetDeployment"
	GetDeploymentsRequest                                       = "GetDeployments"
	GetDomainRequest                                            = "GetDomain"
	GetDomainRouteReservationsRequest                           = "GetDomainRouteReservations"
	GetDomainsRequest                                           = "GetDomains"
	GetDropletRequest                                           = "GetDroplet"
	GetDropletsRequest                                          = "GetDroplets"
	GetEnvironmentVariableGroupRequest                          = "GetEnvironmentVariableGroup"
	GetEventsRequest                                            = "GetEvents"
	GetFeatureFlagRequest                                       = "GetFeatureFlag"
	GetFeatureFlagsRequest                                      = "GetFeatureFlags"
	GetIsolationSegmentOrganizationsRequest                     = "GetIsolationSegmentOrganizations"
	GetIsolationSegmentRequest                                  = "GetIsolationSegment"
	GetIsolationSegmentsRequest                                 = "GetIsolationSegments"
	GetOrganizationDomainsRequest                               = "GetOrganizationDomains"
	GetOrganizationQuotasRequest                                = "GetOrganizationQuotas"
	GetOrganizationQuotaRequest                                 = "GetOrganizationQuota"
	GetOrganizationRelationshipDefaultIsolationSegmentRequest   = "GetOrganizationRelationshipDefaultIsolationSegment"
	GetOrganizationRequest                                      = "GetOrganization"
	GetOrganizationsRequest                                     = "GetOrganizations"
	GetPackageRequest                                           = "GetPackage"
	GetPackagesRequest                                          = "GetPackages"
	GetPackageDropletsRequest                                   = "GetPackageDroplets"
	GetProcessRequest                                           = "GetProcess"
	GetProcessStatsRequest                                      = "GetProcessStats"
	GetProcessSidecarsRequest                                   = "GetProcessSidecars"
	GetRolesRequest                                             = "GetRoles"
	GetRouteDestinationsRequest                                 = "GetRouteDestinations"
	GetRoutesRequest                                            = "GetRoutes"
	GetServiceBrokersRequest                                    = "GetServiceBrokers"
	GetServiceInstancesRequest                                  = "GetServiceInstances"
	GetServiceOfferingsRequest                                  = "GetServiceOfferings"
	GetSpaceRelationshipIsolationSegmentRequest                 = "GetSpaceRelationshipIsolationSegment"
	GetSpacesRequest                                            = "GetSpaces"
	GetSpaceQuotaRequest                                        = "GetSpaceQuota"
	GetSpaceQuotasRequest                                       = "GetSpaceQuotas"
	GetSSHEnabled                                               = "GetSSHEnabled"
	GetStacksRequest                                            = "GetStacks"
	GetUserRequest                                              = "GetUser"
	GetUsersRequest                                             = "GetUsers"
	MapRouteRequest                                             = "MapRoute"
	PatchApplicationCurrentDropletRequest                       = "PatchApplicationCurrentDroplet"
	PatchApplicationEnvironmentVariablesRequest                 = "PatchApplicationEnvironmentVariables"
	PatchApplicationRequest                                     = "PatchApplication"
	PatchApplicationFeaturesRequest                             = "PatchApplicationFeatures"
	PatchEnvironmentVariableGroupRequest                        = "PatchEnvironmentVariableGroup"
	PatchBuildpackRequest                                       = "PatchBuildpack"
	PatchDomainRequest                                          = "PatchDomain"
	PatchFeatureFlagRequest                                     = "PatchFeatureFlag"
	PatchOrganizationRelationshipDefaultIsolationSegmentRequest = "PatchOrganizationRelationshipDefaultIsolationSegment"
	PatchOrganizationRequest                                    = "PatchOrganization"
	PatchOrganizationQuotaRequest                               = "PatchOrganizationQuota"
	PatchProcessRequest                                         = "PatchProcess"
	PatchRouteRequest                                           = "PatchRoute"
	PatchServiceBrokerRequest                                   = "PatchServiceBrokerRequest"
	PatchServiceOfferingRequest                                 = "PatchServiceOfferingRequest"
	PatchSpaceRelationshipIsolationSegmentRequest               = "PatchSpaceRelationshipIsolationSegment"
	PatchSpaceRequest                                           = "PatchSpace"
	PatchSpaceQuotaRequest                                      = "PatchSpaceQuota"
	PatchStackRequest                                           = "PatchStack"
	PostApplicationActionApplyManifest                          = "PostApplicationActionApplyM"
	PostApplicationActionRestartRequest                         = "PostApplicationActionRestart"
	PostApplicationActionStartRequest                           = "PostApplicationActionStart"
	PostApplicationActionStopRequest                            = "PostApplicationActionStop"
	PostApplicationDeploymentActionCancelRequest                = "PostApplicationDeploymentActionCancel"
	PostApplicationDeploymentRequest                            = "PostApplicationDeployment"
	PostApplicationProcessActionScaleRequest                    = "PostApplicationProcessActionScale"
	PostApplicationRequest                                      = "PostApplication"
	PostApplicationTasksRequest                                 = "PostApplicationTasks"
	PostBuildRequest                                            = "PostBuild"
	PostBuildpackBitsRequest                                    = "PostBuildpackBits"
	PostBuildpackRequest                                        = "PostBuildpack"
	PostDomainRequest                                           = "PostDomain"
	PostDropletBitsRequest                                      = "PostDropletBits"
	PostDropletRequest                                          = "PostDroplet"
	PostIsolationSegmentRelationshipOrganizationsRequest        = "PostIsolationSegmentRelationshipOrganizations"
	PostIsolationSegmentsRequest                                = "PostIsolationSegments"
	PostOrganizationRequest                                     = "PostOrganization"
	PostOrganizationQuotaRequest                                = "PostOrganizationQuota"
	PostOrganizationQuotaApplyRequest                           = "PostOrganizationQuotaApply"
	PostPackageRequest                                          = "PostPackage"
	PostPackageBitsRequest                                      = "PostPackageBits"
	PostResourceMatchesRequest                                  = "PostResourceMatches"
	PostRoleRequest                                             = "PostRole"
	PostRouteRequest                                            = "PostRoute"
	PostServiceBrokerRequest                                    = "PostServiceBroker"
	PostServiceInstanceRelationshipsSharedSpacesRequest         = "PostServiceInstanceRelationshipsSharedSpaces"
	PostSpaceActionApplyManifestRequest                         = "PostSpaceActionApplyManifest"
	PostSpaceRequest                                            = "PostSpace"
	PostSpaceQuotaRequest                                       = "PostSpaceQuota"
	PostSpaceQuotaRelationshipsRequest                          = "PostSpaceQuotaRelationships"
	PostUserRequest                                             = "PostUser"
	PutTaskCancelRequest                                        = "PutTaskCancel"
	SharePrivateDomainRequest                                   = "SharePrivateDomainRequest"
	UnmapRouteRequest                                           = "UnmapRoute"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Resource: AppsResource, Path: "/", Method: http.MethodGet, Name: GetApplicationsRequest},
	{Resource: AppsResource, Path: "/", Method: http.MethodPost, Name: PostApplicationRequest},
	{Resource: AppsResource, Path: "/:app_guid", Method: http.MethodDelete, Name: DeleteApplicationRequest},
	{Resource: AppsResource, Path: "/:app_guid", Method: http.MethodPatch, Name: PatchApplicationRequest},
	{Resource: AppsResource, Path: "/:app_guid/features/:name", Method: http.MethodPatch, Name: PatchApplicationFeaturesRequest},
	{Resource: AppsResource, Path: "/:app_guid/features/:name", Method: http.MethodGet, Name: GetApplicationFeaturesRequest},
	{Resource: AppsResource, Path: "/:app_guid/actions/apply_manifest", Method: http.MethodPost, Name: PostApplicationActionApplyManifest},
	{Resource: AppsResource, Path: "/:app_guid/actions/restart", Method: http.MethodPost, Name: PostApplicationActionRestartRequest},
	{Resource: AppsResource, Path: "/:app_guid/actions/start", Method: http.MethodPost, Name: PostApplicationActionStartRequest},
	{Resource: AppsResource, Path: "/:app_guid/actions/stop", Method: http.MethodPost, Name: PostApplicationActionStopRequest},
	{Resource: AppsResource, Path: "/:app_guid/droplets/current", Method: http.MethodGet, Name: GetApplicationDropletCurrentRequest},
	{Resource: AppsResource, Path: "/:app_guid/env", Method: http.MethodGet, Name: GetApplicationEnvRequest},
	{Resource: AppsResource, Path: "/:app_guid/environment_variables", Method: http.MethodPatch, Name: PatchApplicationEnvironmentVariablesRequest},
	{Resource: AppsResource, Path: "/:app_guid/manifest", Method: http.MethodGet, Name: GetApplicationManifestRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes", Method: http.MethodGet, Name: GetApplicationProcessesRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes/:type", Method: http.MethodGet, Name: GetApplicationProcessRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes/:type/actions/scale", Method: http.MethodPost, Name: PostApplicationProcessActionScaleRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes/:type/instances/:index", Method: http.MethodDelete, Name: DeleteApplicationProcessInstanceRequest},
	{Resource: AppsResource, Path: "/:app_guid/relationships/current_droplet", Method: http.MethodPatch, Name: PatchApplicationCurrentDropletRequest},
	{Resource: AppsResource, Path: "/:app_guid/routes", Method: http.MethodGet, Name: GetApplicationRoutesRequest},
	{Resource: AppsResource, Path: "/:app_guid/ssh_enabled", Method: http.MethodGet, Name: GetSSHEnabled},
	{Resource: AppsResource, Path: "/:app_guid/tasks", Method: http.MethodGet, Name: GetApplicationTasksRequest},
	{Resource: AppsResource, Path: "/:app_guid/tasks", Method: http.MethodPost, Name: PostApplicationTasksRequest},
	{Resource: BuildpacksResource, Path: "/", Method: http.MethodGet, Name: GetBuildpacksRequest},
	{Resource: BuildpacksResource, Path: "/", Method: http.MethodPost, Name: PostBuildpackRequest},
	{Resource: BuildpacksResource, Path: "/:buildpack_guid", Method: http.MethodDelete, Name: DeleteBuildpackRequest},
	{Resource: BuildpacksResource, Path: "/:buildpack_guid", Method: http.MethodPatch, Name: PatchBuildpackRequest},
	{Resource: BuildpacksResource, Path: "/:buildpack_guid/upload", Method: http.MethodPost, Name: PostBuildpackBitsRequest},
	{Resource: BuildsResource, Path: "/", Method: http.MethodPost, Name: PostBuildRequest},
	{Resource: BuildsResource, Path: "/:build_guid", Method: http.MethodGet, Name: GetBuildRequest},
	{Resource: DeploymentsResource, Path: "/", Method: http.MethodGet, Name: GetDeploymentsRequest},
	{Resource: DeploymentsResource, Path: "/", Method: http.MethodPost, Name: PostApplicationDeploymentRequest},
	{Resource: DeploymentsResource, Path: "/:deployment_guid", Method: http.MethodGet, Name: GetDeploymentRequest},
	{Resource: DeploymentsResource, Path: "/:deployment_guid/actions/cancel", Method: http.MethodPost, Name: PostApplicationDeploymentActionCancelRequest},
	{Resource: DomainsResource, Path: "/", Method: http.MethodGet, Name: GetDomainsRequest},
	{Resource: DomainsResource, Path: "/", Method: http.MethodPost, Name: PostDomainRequest},
	{Resource: DomainsResource, Path: "/:domain_guid", Method: http.MethodDelete, Name: DeleteDomainRequest},
	{Resource: DomainsResource, Path: "/:domain_guid", Method: http.MethodGet, Name: GetDomainRequest},
	{Resource: DomainsResource, Path: "/:domain_guid", Method: http.MethodPatch, Name: PatchDomainRequest},
	{Resource: DomainsResource, Path: "/:domain_guid/relationships/shared_organizations", Method: http.MethodPost, Name: SharePrivateDomainRequest},
	{Resource: DomainsResource, Path: "/:domain_guid/relationships/shared_organizations/:org_guid", Method: http.MethodDelete, Name: DeleteSharedOrgFromDomainRequest},
	{Resource: DomainsResource, Path: "/:domain_guid/route_reservations", Method: http.MethodGet, Name: GetDomainRouteReservationsRequest},
	{Resource: DropletsResource, Path: "/", Method: http.MethodGet, Name: GetDropletsRequest},
	{Resource: DropletsResource, Path: "/", Method: http.MethodPost, Name: PostDropletRequest},
	{Resource: DropletsResource, Path: "/:droplet_guid", Method: http.MethodGet, Name: GetDropletRequest},
	{Resource: DropletsResource, Path: "/:droplet_guid/upload", Method: http.MethodPost, Name: PostDropletBitsRequest},
	{Resource: EnvironmentVariableGroupsResource, Path: "/:group_name", Method: http.MethodGet, Name: GetEnvironmentVariableGroupRequest},
	{Resource: EnvironmentVariableGroupsResource, Path: "/:group_name", Method: http.MethodPatch, Name: PatchEnvironmentVariableGroupRequest},
	{Resource: EventsResource, Path: "/", Method: http.MethodGet, Name: GetEventsRequest},
	{Resource: FeatureFlagsResource, Path: "/", Method: http.MethodGet, Name: GetFeatureFlagsRequest},
	{Resource: FeatureFlagsResource, Path: "/:name", Method: http.MethodGet, Name: GetFeatureFlagRequest},
	{Resource: FeatureFlagsResource, Path: "/:name", Method: http.MethodPatch, Name: PatchFeatureFlagRequest},
	{Resource: IsolationSegmentsResource, Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest},
	{Resource: IsolationSegmentsResource, Path: "/", Method: http.MethodPost, Name: PostIsolationSegmentsRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid", Method: http.MethodGet, Name: GetIsolationSegmentRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid/organizations", Method: http.MethodGet, Name: GetIsolationSegmentOrganizationsRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid/relationships/organizations", Method: http.MethodPost, Name: PostIsolationSegmentRelationshipOrganizationsRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid/relationships/organizations/:organization_guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRelationshipOrganizationRequest},
	{Resource: OrgsResource, Path: "/", Method: http.MethodGet, Name: GetOrganizationsRequest},
	{Resource: OrgsResource, Path: "/", Method: http.MethodPost, Name: PostOrganizationRequest},
	{Resource: OrgsResource, Path: "/:organization_guid", Method: http.MethodGet, Name: GetOrganizationRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/", Method: http.MethodDelete, Name: DeleteOrganizationRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/", Method: http.MethodPatch, Name: PatchOrganizationRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/domains", Method: http.MethodGet, Name: GetOrganizationDomainsRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/domains/default", Method: http.MethodGet, Name: GetDefaultDomainRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/relationships/default_isolation_segment", Method: http.MethodGet, Name: GetOrganizationRelationshipDefaultIsolationSegmentRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/relationships/default_isolation_segment", Method: http.MethodPatch, Name: PatchOrganizationRelationshipDefaultIsolationSegmentRequest},
	{Resource: OrgQuotasResource, Path: "/:quota_guid", Method: http.MethodPatch, Name: PatchOrganizationQuotaRequest},
	{Resource: OrgQuotasResource, Path: "/", Method: http.MethodPost, Name: PostOrganizationQuotaRequest},
	{Resource: OrgQuotasResource, Path: "/:quota_guid/relationships/organizations", Method: http.MethodPost, Name: PostOrganizationQuotaApplyRequest},
	{Resource: OrgQuotasResource, Path: "/:quota_guid", Method: http.MethodGet, Name: GetOrganizationQuotaRequest},
	{Resource: OrgQuotasResource, Path: "/", Method: http.MethodGet, Name: GetOrganizationQuotasRequest},
	{Resource: OrgQuotasResource, Path: "/:quota_guid", Method: http.MethodDelete, Name: DeleteOrganizationQuotaRequest},
	{Resource: PackagesResource, Path: "/", Method: http.MethodGet, Name: GetPackagesRequest},
	{Resource: PackagesResource, Path: "/", Method: http.MethodPost, Name: PostPackageRequest},
	{Resource: PackagesResource, Path: "/:package_guid", Method: http.MethodGet, Name: GetPackageRequest},
	{Resource: PackagesResource, Path: "/:package_guid/upload", Method: http.MethodPost, Name: PostPackageBitsRequest},
	{Resource: PackagesResource, Path: "/:package_guid/droplets", Method: http.MethodGet, Name: GetPackageDropletsRequest},
	{Resource: ProcessesResource, Path: "/:process_guid", Method: http.MethodGet, Name: GetProcessRequest},
	{Resource: ProcessesResource, Path: "/:process_guid", Method: http.MethodPatch, Name: PatchProcessRequest},
	{Resource: ProcessesResource, Path: "/:process_guid/stats", Method: http.MethodGet, Name: GetProcessStatsRequest},
	{Resource: ProcessesResource, Path: "/:process_guid/sidecars", Method: http.MethodGet, Name: GetProcessSidecarsRequest},
	{Resource: ResourceMatches, Path: "/", Method: http.MethodPost, Name: PostResourceMatchesRequest},
	{Resource: RolesResource, Path: "/", Method: http.MethodGet, Name: GetRolesRequest},
	{Resource: RolesResource, Path: "/", Method: http.MethodPost, Name: PostRoleRequest},
	{Resource: RoutesResource, Path: "/", Method: http.MethodGet, Name: GetRoutesRequest},
	{Resource: RoutesResource, Path: "/", Method: http.MethodPost, Name: PostRouteRequest},
	{Resource: RoutesResource, Path: "/:route_guid", Method: http.MethodDelete, Name: DeleteRouteRequest},
	{Resource: RoutesResource, Path: "/:route_guid", Method: http.MethodPatch, Name: PatchRouteRequest},
	{Resource: RoutesResource, Path: "/:route_guid/destinations", Method: http.MethodGet, Name: GetRouteDestinationsRequest},
	{Resource: RoutesResource, Path: "/:route_guid/destinations", Method: http.MethodPost, Name: MapRouteRequest},
	{Resource: RoutesResource, Path: "/:route_guid/destinations/:destination_guid", Method: http.MethodDelete, Name: UnmapRouteRequest},
	{Resource: RolesResource, Path: "/:role_guid", Method: http.MethodDelete, Name: DeleteRoleRequest},
	{Resource: ServiceBrokersResource, Path: "/", Method: http.MethodGet, Name: GetServiceBrokersRequest},
	{Resource: ServiceBrokersResource, Path: "/", Method: http.MethodPost, Name: PostServiceBrokerRequest},
	{Resource: ServiceBrokersResource, Path: "/:service_broker_guid", Method: http.MethodDelete, Name: DeleteServiceBrokerRequest},
	{Resource: ServiceBrokersResource, Path: "/:service_broker_guid", Method: http.MethodPatch, Name: PatchServiceBrokerRequest},
	{Resource: ServiceInstancesResource, Path: "/", Method: http.MethodGet, Name: GetServiceInstancesRequest},
	{Resource: ServiceInstancesResource, Path: "/:service_instance_guid/relationships/shared_spaces", Method: http.MethodPost, Name: PostServiceInstanceRelationshipsSharedSpacesRequest},
	{Resource: ServiceInstancesResource, Path: "/:service_instance_guid/relationships/shared_spaces/:space_guid", Method: http.MethodDelete, Name: DeleteServiceInstanceRelationshipsSharedSpaceRequest},
	{Resource: ServiceOfferingsResource, Path: "/", Method: http.MethodGet, Name: GetServiceOfferingsRequest},
	{Resource: ServiceOfferingsResource, Path: "/:service_offering_guid", Method: http.MethodPatch, Name: PatchServiceOfferingRequest},
	{Resource: SpacesResource, Path: "/", Method: http.MethodGet, Name: GetSpacesRequest},
	{Resource: SpacesResource, Path: "/", Method: http.MethodPost, Name: PostSpaceRequest},
	{Resource: SpacesResource, Path: "/:space_guid", Method: http.MethodDelete, Name: DeleteSpaceRequest},
	{Resource: SpacesResource, Path: "/:space_guid", Method: http.MethodPatch, Name: PatchSpaceRequest},
	{Resource: SpacesResource, Path: "/:space_guid/actions/apply_manifest", Method: http.MethodPost, Name: PostSpaceActionApplyManifestRequest},
	{Resource: SpacesResource, Path: "/:space_guid/relationships/isolation_segment", Method: http.MethodGet, Name: GetSpaceRelationshipIsolationSegmentRequest},
	{Resource: SpacesResource, Path: "/:space_guid/relationships/isolation_segment", Method: http.MethodPatch, Name: PatchSpaceRelationshipIsolationSegmentRequest},
	{Resource: SpacesResource, Path: "/:space_guid/routes", Method: http.MethodDelete, Name: DeleteOrphanedRoutesRequest},
	{Resource: SpaceQuotasResource, Path: "/", Method: http.MethodPost, Name: PostSpaceQuotaRequest},
	{Resource: SpaceQuotasResource, Path: "/:quota_guid", Method: http.MethodGet, Name: GetSpaceQuotaRequest},
	{Resource: SpaceQuotasResource, Path: "/:quota_guid", Method: http.MethodDelete, Name: DeleteSpaceQuotaRequest},
	{Resource: SpaceQuotasResource, Path: "/:quota_guid/relationships/spaces", Method: http.MethodPost, Name: PostSpaceQuotaRelationshipsRequest},
	{Resource: SpaceQuotasResource, Path: "/", Method: http.MethodGet, Name: GetSpaceQuotasRequest},
	{Resource: SpaceQuotasResource, Path: "/:quota_guid", Method: http.MethodPatch, Name: PatchSpaceQuotaRequest},
	{Resource: SpaceQuotasResource, Path: "/:quota_guid/relationships/spaces/:space_guid", Method: http.MethodDelete, Name: DeleteSpaceQuotaFromSpaceRequest},
	{Resource: StacksResource, Path: "/", Method: http.MethodGet, Name: GetStacksRequest},
	{Resource: StacksResource, Path: "/:stack_guid", Method: http.MethodPatch, Name: PatchStackRequest},
	{Resource: TasksResource, Path: "/:task_guid/cancel", Method: http.MethodPut, Name: PutTaskCancelRequest},
	{Resource: UsersResource, Path: "/", Method: http.MethodGet, Name: GetUsersRequest},
	{Resource: UsersResource, Path: "/:user_guid", Method: http.MethodGet, Name: GetUserRequest},
	{Resource: UsersResource, Path: "/", Method: http.MethodPost, Name: PostUserRequest},
	{Resource: UsersResource, Path: "/:user_guid", Method: http.MethodDelete, Name: DeleteUserRequest},
}
