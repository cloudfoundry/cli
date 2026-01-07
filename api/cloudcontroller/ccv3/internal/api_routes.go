package internal

import "net/http"

// Naming convention:
//
// HTTP method + non-parameter parts of the path + "Request"
//
// If the request returns a single entity by GUID, use the singular (for example
// /v3/organizations/:organization_guid is GetOrganization).
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
	DeleteRoleRequest                                           = "DeleteRoleRequest"
	DeleteRouteRequest                                          = "DeleteRouteRequest"
	DeleteRouteBindingRequest                                   = "DeleteRouteBinding"
	DeleteSecurityGroupRequest                                  = "DeleteSecurityGroup"
	DeleteSecurityGroupStagingSpaceRequest                      = "DeleteSecurityGroupStagingSpace"
	DeleteSecurityGroupRunningSpaceRequest                      = "DeleteSecurityGroupRunningSpace"
	DeleteServiceCredentialBindingRequest                       = "DeleteServiceCredentialBinding"
	DeleteServiceBrokerRequest                                  = "DeleteServiceBrokerRequest"
	DeleteServiceInstanceRelationshipsSharedSpaceRequest        = "DeleteServiceInstanceRelationshipsSharedSpace"
	DeleteServiceInstanceRequest                                = "DeleteServiceInstance"
	DeleteServiceOfferingRequest                                = "DeleteServiceOffering"
	DeleteServicePlanVisibilityRequest                          = "DeleteServicePlanVisibility"
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
	GetApplicationRevisionsRequest                              = "GetApplicationRevisions"
	GetApplicationRevisionsDeployedRequest                      = "GetApplicationRevisionsDeployed"
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
	GetDropletBitsRequest                                       = "GetDropletBits"
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
	GetProcessesRequest                                         = "GetProcesses"
	GetProcessStatsRequest                                      = "GetProcessStats"
	GetProcessSidecarsRequest                                   = "GetProcessSidecars"
	GetRolesRequest                                             = "GetRoles"
	GetRouteBindingsRequest                                     = "GetRouteBindings"
	GetRouteDestinationsRequest                                 = "GetRouteDestinations"
	GetRoutesRequest                                            = "GetRoutes"
	GetSecurityGroupsRequest                                    = "GetSecurityGroups"
	GetServiceBrokersRequest                                    = "GetServiceBrokers"
	GetServiceCredentialBindingsRequest                         = "GetServiceCredentialBindings"
	GetServiceCredentialBindingDetailsRequest                   = "GetServiceCredentialBindingDetails"
	GetServiceInstanceParametersRequest                         = "GetServiceInstanceParameters"
	GetServiceInstanceRequest                                   = "GetServiceInstance"
	GetServiceInstancesRequest                                  = "GetServiceInstances"
	GetServiceInstanceRelationshipsSharedSpacesRequest          = "GetServiceInstanceRelationshipSharedSpacesRequest"
	GetServiceInstanceSharedSpacesUsageSummaryRequest           = "GetServiceInstanceSharedSpacesUsageSummaryRequest"
	GetServiceOfferingRequest                                   = "GetServiceOffering"
	GetServiceOfferingsRequest                                  = "GetServiceOfferings"
	GetServicePlanRequest                                       = "GetServicePlan"
	GetServicePlansRequest                                      = "GetServicePlans"
	GetServicePlanVisibilityRequest                             = "GetServicePlanVisibility"
	GetSpaceFeatureRequest                                      = "GetSpaceFeatureRequest"
	GetSpaceRelationshipIsolationSegmentRequest                 = "GetSpaceRelationshipIsolationSegment"
	GetSpaceRunningSecurityGroupsRequest                        = "GetSpaceRunningSecurityGroups"
	GetSpacesRequest                                            = "GetSpaces"
	GetSpaceQuotaRequest                                        = "GetSpaceQuota"
	GetSpaceQuotasRequest                                       = "GetSpaceQuotas"
	GetSpaceStagingSecurityGroupsRequest                        = "GetSpaceStagingSecurityGroups"
	GetSSHEnabled                                               = "GetSSHEnabled"
	GetStacksRequest                                            = "GetStacks"
	GetTasksRequest                                             = "GetTasks"
	GetTaskRequest                                              = "GetTask"
	GetUserRequest                                              = "GetUser"
	GetUsersRequest                                             = "GetUsers"
	MapRouteRequest                                             = "MapRoute"
	PatchApplicationCurrentDropletRequest                       = "PatchApplicationCurrentDroplet"
	PatchApplicationEnvironmentVariablesRequest                 = "PatchApplicationEnvironmentVariables"
	PatchApplicationRequest                                     = "PatchApplication"
	PatchApplicationFeaturesRequest                             = "PatchApplicationFeatures"
	PatchEnvironmentVariableGroupRequest                        = "PatchEnvironmentVariableGroup"
	PatchBuildpackRequest                                       = "PatchBuildpack"
	PatchDestinationRequest                                     = "PatchDestination"
	PatchDomainRequest                                          = "PatchDomain"
	PatchFeatureFlagRequest                                     = "PatchFeatureFlag"
	PatchOrganizationRelationshipDefaultIsolationSegmentRequest = "PatchOrganizationRelationshipDefaultIsolationSegment"
	PatchOrganizationRequest                                    = "PatchOrganization"
	PatchOrganizationQuotaRequest                               = "PatchOrganizationQuota"
	PatchProcessRequest                                         = "PatchProcess"
	PatchRouteRequest                                           = "PatchRoute"
	PatchSecurityGroupRequest                                   = "PatchSecurityGroup"
	PatchServiceBrokerRequest                                   = "PatchServiceBrokerRequest"
	PatchServiceInstanceRequest                                 = "PatchServiceInstance"
	PatchServiceOfferingRequest                                 = "PatchServiceOfferingRequest"
	PatchServicePlanRequest                                     = "PatchServicePlanRequest"
	PatchSpaceRelationshipIsolationSegmentRequest               = "PatchSpaceRelationshipIsolationSegment"
	PatchSpaceRequest                                           = "PatchSpace"
	PatchSpaceFeaturesRequest                                   = "PatchSpaceFeatures"
	PatchSpaceQuotaRequest                                      = "PatchSpaceQuota"
	PatchStackRequest                                           = "PatchStack"
	PatchMoveRouteRequest                                       = "PatchMoveRouteRequest"
	PostApplicationActionApplyManifest                          = "PostApplicationActionApplyM"
	PostApplicationActionRestartRequest                         = "PostApplicationActionRestart"
	PostApplicationActionStartRequest                           = "PostApplicationActionStart"
	PostApplicationActionStopRequest                            = "PostApplicationActionStop"
	PostApplicationDeploymentActionCancelRequest                = "PostApplicationDeploymentActionCancel"
	PostApplicationDeploymentActionContinueRequest              = "PostApplicationDeploymentActionContinue"
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
	PostRouteBindingRequest                                     = "PostRouteBinding"
	PostSecurityGroupRequest                                    = "PostSecurityGroup"
	PostSecurityGroupStagingSpaceRequest                        = "PostSecurityGroupStagingSpace"
	PostSecurityGroupRunningSpaceRequest                        = "PostSecurityGroupRunningSpace"
	PostServiceCredentialBindingRequest                         = "PostServiceCredentialBinding"
	PostServiceBrokerRequest                                    = "PostServiceBroker"
	PostServiceInstanceRequest                                  = "PostServiceInstance"
	PostServiceInstanceRelationshipsSharedSpacesRequest         = "PostServiceInstanceRelationshipsSharedSpaces"
	PostServicePlanVisibilityRequest                            = "PostServicePlanVisibility"
	PostSpaceActionApplyManifestRequest                         = "PostSpaceActionApplyManifest"
	PostSpaceDiffManifestRequest                                = "PostSpaceDiffManifest"
	PostSpaceRequest                                            = "PostSpace"
	PostSpaceQuotaRequest                                       = "PostSpaceQuota"
	PostSpaceQuotaRelationshipsRequest                          = "PostSpaceQuotaRelationships"
	PostUserRequest                                             = "PostUser"
	PutTaskCancelRequest                                        = "PutTaskCancel"
	SharePrivateDomainRequest                                   = "SharePrivateDomainRequest"
	ShareRouteRequest                                           = "ShareRouteRequest"
	UnmapRouteRequest                                           = "UnmapRoute"
	UnshareRouteRequest                                         = "UnshareRoute"
	UpdateRouteRequest                                          = "UpdateRoute"
	WhoAmI                                                      = "WhoAmI"
	Info                                                        = "Info"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = map[string]Route{
	GetApplicationsRequest:                                      {Path: "/v3/apps", Method: http.MethodGet},
	PostApplicationRequest:                                      {Path: "/v3/apps", Method: http.MethodPost},
	DeleteApplicationRequest:                                    {Path: "/v3/apps/:app_guid", Method: http.MethodDelete},
	PatchApplicationRequest:                                     {Path: "/v3/apps/:app_guid", Method: http.MethodPatch},
	PatchApplicationFeaturesRequest:                             {Path: "/v3/apps/:app_guid/features/:name", Method: http.MethodPatch},
	GetApplicationFeaturesRequest:                               {Path: "/v3/apps/:app_guid/features/:name", Method: http.MethodGet},
	PostApplicationActionApplyManifest:                          {Path: "/v3/apps/:app_guid/actions/apply_manifest", Method: http.MethodPost},
	PostApplicationActionRestartRequest:                         {Path: "/v3/apps/:app_guid/actions/restart", Method: http.MethodPost},
	PostApplicationActionStartRequest:                           {Path: "/v3/apps/:app_guid/actions/start", Method: http.MethodPost},
	PostApplicationActionStopRequest:                            {Path: "/v3/apps/:app_guid/actions/stop", Method: http.MethodPost},
	GetApplicationDropletCurrentRequest:                         {Path: "/v3/apps/:app_guid/droplets/current", Method: http.MethodGet},
	GetApplicationEnvRequest:                                    {Path: "/v3/apps/:app_guid/env", Method: http.MethodGet},
	PatchApplicationEnvironmentVariablesRequest:                 {Path: "/v3/apps/:app_guid/environment_variables", Method: http.MethodPatch},
	GetApplicationManifestRequest:                               {Path: "/v3/apps/:app_guid/manifest", Method: http.MethodGet},
	GetApplicationProcessesRequest:                              {Path: "/v3/apps/:app_guid/processes", Method: http.MethodGet},
	GetApplicationProcessRequest:                                {Path: "/v3/apps/:app_guid/processes/:type", Method: http.MethodGet},
	PostApplicationProcessActionScaleRequest:                    {Path: "/v3/apps/:app_guid/processes/:type/actions/scale", Method: http.MethodPost},
	DeleteApplicationProcessInstanceRequest:                     {Path: "/v3/apps/:app_guid/processes/:type/instances/:index", Method: http.MethodDelete},
	PatchApplicationCurrentDropletRequest:                       {Path: "/v3/apps/:app_guid/relationships/current_droplet", Method: http.MethodPatch},
	GetApplicationRevisionsRequest:                              {Path: "/v3/apps/:app_guid/revisions", Method: http.MethodGet},
	GetApplicationRevisionsDeployedRequest:                      {Path: "/v3/apps/:app_guid/revisions/deployed", Method: http.MethodGet},
	GetApplicationRoutesRequest:                                 {Path: "/v3/apps/:app_guid/routes", Method: http.MethodGet},
	GetSSHEnabled:                                               {Path: "/v3/apps/:app_guid/ssh_enabled", Method: http.MethodGet},
	GetApplicationTasksRequest:                                  {Path: "/v3/apps/:app_guid/tasks", Method: http.MethodGet},
	PostApplicationTasksRequest:                                 {Path: "/v3/apps/:app_guid/tasks", Method: http.MethodPost},
	GetBuildpacksRequest:                                        {Path: "/v3/buildpacks", Method: http.MethodGet},
	PostBuildpackRequest:                                        {Path: "/v3/buildpacks/", Method: http.MethodPost},
	DeleteBuildpackRequest:                                      {Path: "/v3/buildpacks/:buildpack_guid", Method: http.MethodDelete},
	PatchBuildpackRequest:                                       {Path: "/v3/buildpacks/:buildpack_guid", Method: http.MethodPatch},
	PostBuildpackBitsRequest:                                    {Path: "/v3/buildpacks/:buildpack_guid/upload", Method: http.MethodPost},
	PostBuildRequest:                                            {Path: "/v3/builds", Method: http.MethodPost},
	GetBuildRequest:                                             {Path: "/v3/builds/:build_guid", Method: http.MethodGet},
	GetDeploymentsRequest:                                       {Path: "/v3/deployments", Method: http.MethodGet},
	PostApplicationDeploymentRequest:                            {Path: "/v3/deployments", Method: http.MethodPost},
	GetDeploymentRequest:                                        {Path: "/v3/deployments/:deployment_guid", Method: http.MethodGet},
	PostApplicationDeploymentActionCancelRequest:                {Path: "/v3/deployments/:deployment_guid/actions/cancel", Method: http.MethodPost},
	PostApplicationDeploymentActionContinueRequest:              {Path: "/v3/deployments/:deployment_guid/actions/continue", Method: http.MethodPost},
	GetDomainsRequest:                                           {Path: "/v3/domains", Method: http.MethodGet},
	PostDomainRequest:                                           {Path: "/v3/domains", Method: http.MethodPost},
	DeleteDomainRequest:                                         {Path: "/v3/domains/:domain_guid", Method: http.MethodDelete},
	GetDomainRequest:                                            {Path: "/v3/domains/:domain_guid", Method: http.MethodGet},
	PatchDomainRequest:                                          {Path: "/v3/domains/:domain_guid", Method: http.MethodPatch},
	SharePrivateDomainRequest:                                   {Path: "/v3/domains/:domain_guid/relationships/shared_organizations", Method: http.MethodPost},
	DeleteSharedOrgFromDomainRequest:                            {Path: "/v3/domains/:domain_guid/relationships/shared_organizations/:org_guid", Method: http.MethodDelete},
	GetDomainRouteReservationsRequest:                           {Path: "/v3/domains/:domain_guid/route_reservations", Method: http.MethodGet},
	GetDropletsRequest:                                          {Path: "/v3/droplets", Method: http.MethodGet},
	PostDropletRequest:                                          {Path: "/v3/droplets", Method: http.MethodPost},
	GetDropletRequest:                                           {Path: "/v3/droplets/:droplet_guid", Method: http.MethodGet},
	PostDropletBitsRequest:                                      {Path: "/v3/droplets/:droplet_guid/upload", Method: http.MethodPost},
	GetDropletBitsRequest:                                       {Path: "/v3/droplets/:droplet_guid/download", Method: http.MethodGet},
	GetEnvironmentVariableGroupRequest:                          {Path: "/v3/environment_variable_groups/:group_name", Method: http.MethodGet},
	PatchEnvironmentVariableGroupRequest:                        {Path: "/v3/environment_variable_groups/:group_name", Method: http.MethodPatch},
	GetEventsRequest:                                            {Path: "/v3/audit_events", Method: http.MethodGet},
	GetFeatureFlagsRequest:                                      {Path: "/v3/feature_flags", Method: http.MethodGet},
	GetFeatureFlagRequest:                                       {Path: "/v3/feature_flags/:name", Method: http.MethodGet},
	PatchFeatureFlagRequest:                                     {Path: "/v3/feature_flags/:name", Method: http.MethodPatch},
	GetIsolationSegmentsRequest:                                 {Path: "/v3/isolation_segments", Method: http.MethodGet},
	PostIsolationSegmentsRequest:                                {Path: "/v3/isolation_segments", Method: http.MethodPost},
	DeleteIsolationSegmentRequest:                               {Path: "/v3/isolation_segments/:isolation_segment_guid", Method: http.MethodDelete},
	GetIsolationSegmentRequest:                                  {Path: "/v3/isolation_segments/:isolation_segment_guid", Method: http.MethodGet},
	GetIsolationSegmentOrganizationsRequest:                     {Path: "/v3/isolation_segments/:isolation_segment_guid/organizations", Method: http.MethodGet},
	PostIsolationSegmentRelationshipOrganizationsRequest:        {Path: "/v3/isolation_segments/:isolation_segment_guid/relationships/organizations", Method: http.MethodPost},
	DeleteIsolationSegmentRelationshipOrganizationRequest:       {Path: "/v3/isolation_segments/:isolation_segment_guid/relationships/organizations/:organization_guid", Method: http.MethodDelete},
	GetOrganizationsRequest:                                     {Path: "/v3/organizations", Method: http.MethodGet},
	PostOrganizationRequest:                                     {Path: "/v3/organizations", Method: http.MethodPost},
	GetOrganizationRequest:                                      {Path: "/v3/organizations/:organization_guid", Method: http.MethodGet},
	DeleteOrganizationRequest:                                   {Path: "/v3/organizations/:organization_guid/", Method: http.MethodDelete},
	PatchOrganizationRequest:                                    {Path: "/v3/organizations/:organization_guid/", Method: http.MethodPatch},
	GetOrganizationDomainsRequest:                               {Path: "/v3/organizations/:organization_guid/domains", Method: http.MethodGet},
	GetDefaultDomainRequest:                                     {Path: "/v3/organizations/:organization_guid/domains/default", Method: http.MethodGet},
	GetOrganizationRelationshipDefaultIsolationSegmentRequest:   {Path: "/v3/organizations/:organization_guid/relationships/default_isolation_segment", Method: http.MethodGet},
	PatchOrganizationRelationshipDefaultIsolationSegmentRequest: {Path: "/v3/organizations/:organization_guid/relationships/default_isolation_segment", Method: http.MethodPatch},
	PatchOrganizationQuotaRequest:                               {Path: "/v3/organization_quotas/:quota_guid", Method: http.MethodPatch},
	PostOrganizationQuotaRequest:                                {Path: "/v3/organization_quotas", Method: http.MethodPost},
	PostOrganizationQuotaApplyRequest:                           {Path: "/v3/organization_quotas/:quota_guid/relationships/organizations", Method: http.MethodPost},
	GetOrganizationQuotaRequest:                                 {Path: "/v3/organization_quotas/:quota_guid", Method: http.MethodGet},
	GetOrganizationQuotasRequest:                                {Path: "/v3/organization_quotas", Method: http.MethodGet},
	DeleteOrganizationQuotaRequest:                              {Path: "/v3/organization_quotas/:quota_guid", Method: http.MethodDelete},
	GetPackagesRequest:                                          {Path: "/v3/packages", Method: http.MethodGet},
	PostPackageRequest:                                          {Path: "/v3/packages", Method: http.MethodPost},
	GetPackageRequest:                                           {Path: "/v3/packages/:package_guid", Method: http.MethodGet},
	PostPackageBitsRequest:                                      {Path: "/v3/packages/:package_guid/upload", Method: http.MethodPost},
	GetPackageDropletsRequest:                                   {Path: "/v3/packages/:package_guid/droplets", Method: http.MethodGet},
	GetProcessRequest:                                           {Path: "/v3/processes/:process_guid", Method: http.MethodGet},
	GetProcessesRequest:                                         {Path: "/v3/processes", Method: http.MethodGet},
	PatchProcessRequest:                                         {Path: "/v3/processes/:process_guid", Method: http.MethodPatch},
	GetProcessStatsRequest:                                      {Path: "/v3/processes/:process_guid/stats", Method: http.MethodGet},
	GetProcessSidecarsRequest:                                   {Path: "/v3/processes/:process_guid/sidecars", Method: http.MethodGet},
	PostResourceMatchesRequest:                                  {Path: "/v3/resource_matches", Method: http.MethodPost},
	GetRolesRequest:                                             {Path: "/v3/roles", Method: http.MethodGet},
	PostRoleRequest:                                             {Path: "/v3/roles", Method: http.MethodPost},
	DeleteRoleRequest:                                           {Path: "/v3/roles/:role_guid", Method: http.MethodDelete},
	GetRoutesRequest:                                            {Path: "/v3/routes", Method: http.MethodGet},
	PostRouteRequest:                                            {Path: "/v3/routes", Method: http.MethodPost},
	DeleteRouteRequest:                                          {Path: "/v3/routes/:route_guid", Method: http.MethodDelete},
	PatchRouteRequest:                                           {Path: "/v3/routes/:route_guid", Method: http.MethodPatch},
	GetRouteDestinationsRequest:                                 {Path: "/v3/routes/:route_guid/destinations", Method: http.MethodGet},
	MapRouteRequest:                                             {Path: "/v3/routes/:route_guid/destinations", Method: http.MethodPost},
	UpdateRouteRequest:                                          {Path: "/v3/routes/:route_guid", Method: http.MethodPatch},
	UnmapRouteRequest:                                           {Path: "/v3/routes/:route_guid/destinations/:destination_guid", Method: http.MethodDelete},
	PatchDestinationRequest:                                     {Path: "/v3/routes/:route_guid/destinations/:destination_guid", Method: http.MethodPatch},
	ShareRouteRequest:                                           {Path: "/v3/routes/:route_guid/relationships/shared_spaces", Method: http.MethodPost},
	UnshareRouteRequest:                                         {Path: "/v3/routes/:route_guid/relationships/shared_spaces/:space_guid", Method: http.MethodDelete},
	PatchMoveRouteRequest:                                       {Path: "/v3/routes/:route_guid/relationships/space", Method: http.MethodPatch},
	GetSecurityGroupsRequest:                                    {Path: "/v3/security_groups", Method: http.MethodGet},
	PostSecurityGroupRequest:                                    {Path: "/v3/security_groups", Method: http.MethodPost},
	DeleteSecurityGroupRequest:                                  {Path: "/v3/security_groups/:security_group_guid", Method: http.MethodDelete},
	PostSecurityGroupStagingSpaceRequest:                        {Path: "/v3/security_groups/:security_group_guid/relationships/staging_spaces", Method: http.MethodPost},
	PostSecurityGroupRunningSpaceRequest:                        {Path: "/v3/security_groups/:security_group_guid/relationships/running_spaces", Method: http.MethodPost},
	DeleteSecurityGroupStagingSpaceRequest:                      {Path: "/v3/security_groups/:security_group_guid/relationships/staging_spaces/:space_guid", Method: http.MethodDelete},
	DeleteSecurityGroupRunningSpaceRequest:                      {Path: "/v3/security_groups/:security_group_guid/relationships/running_spaces/:space_guid", Method: http.MethodDelete},
	PatchSecurityGroupRequest:                                   {Path: "/v3/security_groups/:security_group_guid", Method: http.MethodPatch},
	GetServiceBrokersRequest:                                    {Path: "/v3/service_brokers", Method: http.MethodGet},
	PostServiceBrokerRequest:                                    {Path: "/v3/service_brokers", Method: http.MethodPost},
	DeleteServiceBrokerRequest:                                  {Path: "/v3/service_brokers/:service_broker_guid", Method: http.MethodDelete},
	PatchServiceBrokerRequest:                                   {Path: "/v3/service_brokers/:service_broker_guid", Method: http.MethodPatch},
	PostServiceCredentialBindingRequest:                         {Path: "/v3/service_credential_bindings", Method: http.MethodPost},
	GetServiceCredentialBindingsRequest:                         {Path: "/v3/service_credential_bindings", Method: http.MethodGet},
	DeleteServiceCredentialBindingRequest:                       {Path: "/v3/service_credential_bindings/:service_credential_binding_guid", Method: http.MethodDelete},
	GetServiceCredentialBindingDetailsRequest:                   {Path: "/v3/service_credential_bindings/:service_credential_binding_guid/details", Method: http.MethodGet},
	GetServiceInstanceRequest:                                   {Path: "/v3/service_instances/:service_instance_guid", Method: http.MethodGet},
	GetServiceInstancesRequest:                                  {Path: "/v3/service_instances", Method: http.MethodGet},
	PostServiceInstanceRequest:                                  {Path: "/v3/service_instances", Method: http.MethodPost},
	GetServiceInstanceParametersRequest:                         {Path: "/v3/service_instances/:service_instance_guid/parameters", Method: http.MethodGet},
	PatchServiceInstanceRequest:                                 {Path: "/v3/service_instances/:service_instance_guid", Method: http.MethodPatch},
	DeleteServiceInstanceRequest:                                {Path: "/v3/service_instances/:service_instance_guid", Method: http.MethodDelete},
	GetServiceInstanceSharedSpacesUsageSummaryRequest:           {Path: "/v3/service_instances/:service_instance_guid/relationships/shared_spaces/usage_summary", Method: http.MethodGet},
	GetServiceInstanceRelationshipsSharedSpacesRequest:          {Path: "/v3/service_instances/:service_instance_guid/relationships/shared_spaces", Method: http.MethodGet},
	PostServiceInstanceRelationshipsSharedSpacesRequest:         {Path: "/v3/service_instances/:service_instance_guid/relationships/shared_spaces", Method: http.MethodPost},
	DeleteServiceInstanceRelationshipsSharedSpaceRequest:        {Path: "/v3/service_instances/:service_instance_guid/relationships/shared_spaces/:space_guid", Method: http.MethodDelete},
	GetServiceOfferingRequest:                                   {Path: "/v3/service_offerings/:service_offering_guid", Method: http.MethodGet},
	GetServiceOfferingsRequest:                                  {Path: "/v3/service_offerings", Method: http.MethodGet},
	PatchServiceOfferingRequest:                                 {Path: "/v3/service_offerings/:service_offering_guid", Method: http.MethodPatch},
	DeleteServiceOfferingRequest:                                {Path: "/v3/service_offerings/:service_offering_guid", Method: http.MethodDelete},
	GetServicePlanRequest:                                       {Path: "/v3/service_plans/:service_plan_guid", Method: http.MethodGet},
	GetServicePlansRequest:                                      {Path: "/v3/service_plans", Method: http.MethodGet},
	PatchServicePlanRequest:                                     {Path: "/v3/service_plans/:service_plan_guid", Method: http.MethodPatch},
	GetServicePlanVisibilityRequest:                             {Path: "/v3/service_plans/:service_plan_guid/visibility", Method: http.MethodGet},
	PostServicePlanVisibilityRequest:                            {Path: "/v3/service_plans/:service_plan_guid/visibility", Method: http.MethodPost},
	DeleteServicePlanVisibilityRequest:                          {Path: "/v3/service_plans/:service_plan_guid/visibility/:organization_guid", Method: http.MethodDelete},
	PostRouteBindingRequest:                                     {Path: "/v3/service_route_bindings", Method: http.MethodPost},
	GetRouteBindingsRequest:                                     {Path: "/v3/service_route_bindings", Method: http.MethodGet},
	DeleteRouteBindingRequest:                                   {Path: "/v3/service_route_bindings/:route_binding_guid", Method: http.MethodDelete},
	GetSpacesRequest:                                            {Path: "/v3/spaces", Method: http.MethodGet},
	PostSpaceRequest:                                            {Path: "/v3/spaces", Method: http.MethodPost},
	DeleteSpaceRequest:                                          {Path: "/v3/spaces/:space_guid", Method: http.MethodDelete},
	PatchSpaceRequest:                                           {Path: "/v3/spaces/:space_guid", Method: http.MethodPatch},
	PostSpaceActionApplyManifestRequest:                         {Path: "/v3/spaces/:space_guid/actions/apply_manifest", Method: http.MethodPost},
	PostSpaceDiffManifestRequest:                                {Path: "/v3/spaces/:space_guid/manifest_diff", Method: http.MethodPost},
	GetSpaceRelationshipIsolationSegmentRequest:                 {Path: "/v3/spaces/:space_guid/relationships/isolation_segment", Method: http.MethodGet},
	PatchSpaceRelationshipIsolationSegmentRequest:               {Path: "/v3/spaces/:space_guid/relationships/isolation_segment", Method: http.MethodPatch},
	DeleteOrphanedRoutesRequest:                                 {Path: "/v3/spaces/:space_guid/routes", Method: http.MethodDelete},
	GetSpaceRunningSecurityGroupsRequest:                        {Path: "/v3/spaces/:space_guid/running_security_groups", Method: http.MethodGet},
	GetSpaceStagingSecurityGroupsRequest:                        {Path: "/v3/spaces/:space_guid/staging_security_groups", Method: http.MethodGet},
	PatchSpaceFeaturesRequest:                                   {Path: "/v3/spaces/:space_guid/features/:feature", Method: http.MethodPatch},
	GetSpaceFeatureRequest:                                      {Path: "/v3/spaces/:space_guid/features/:feature", Method: http.MethodGet},
	PostSpaceQuotaRequest:                                       {Path: "/v3/space_quotas", Method: http.MethodPost},
	GetSpaceQuotaRequest:                                        {Path: "/v3/space_quotas/:quota_guid", Method: http.MethodGet},
	DeleteSpaceQuotaRequest:                                     {Path: "/v3/space_quotas/:quota_guid", Method: http.MethodDelete},
	PostSpaceQuotaRelationshipsRequest:                          {Path: "/v3/space_quotas/:quota_guid/relationships/spaces", Method: http.MethodPost},
	GetSpaceQuotasRequest:                                       {Path: "/v3/space_quotas", Method: http.MethodGet},
	PatchSpaceQuotaRequest:                                      {Path: "/v3/space_quotas/:quota_guid", Method: http.MethodPatch},
	DeleteSpaceQuotaFromSpaceRequest:                            {Path: "/v3/space_quotas/:quota_guid/relationships/spaces/:space_guid", Method: http.MethodDelete},
	GetStacksRequest:                                            {Path: "/v3/stacks", Method: http.MethodGet},
	PatchStackRequest:                                           {Path: "/v3/stacks/:stack_guid", Method: http.MethodPatch},
	GetTaskRequest:                                              {Path: "/v3/tasks/:task_guid", Method: http.MethodGet},
	PutTaskCancelRequest:                                        {Path: "/v3/tasks/:task_guid/cancel", Method: http.MethodPut},
	GetTasksRequest:                                             {Path: "/v3/tasks", Method: http.MethodGet},
	GetUsersRequest:                                             {Path: "/v3/users", Method: http.MethodGet},
	GetUserRequest:                                              {Path: "/v3/users/:user_guid", Method: http.MethodGet},
	PostUserRequest:                                             {Path: "/v3/users", Method: http.MethodPost},
	DeleteUserRequest:                                           {Path: "/v3/users/:user_guid", Method: http.MethodDelete},
	WhoAmI:                                                      {Path: "/whoami", Method: http.MethodGet},
	Info:                                                        {Path: "/v3/info", Method: http.MethodGet},
}
