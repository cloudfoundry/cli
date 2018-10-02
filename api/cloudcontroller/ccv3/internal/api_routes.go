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
	DeleteIsolationSegmentRelationshipOrganizationRequest       = "DeleteIsolationSegmentRelationshipOrganization"
	DeleteIsolationSegmentRequest                               = "DeleteIsolationSegment"
	DeleteServiceInstanceRelationshipsSharedSpaceRequest        = "DeleteServiceInstanceRelationshipsSharedSpace"
	GetApplicationDropletCurrentRequest                         = "GetApplicationDropletCurrent"
	GetApplicationEnvRequest                                    = "GetApplicationEnv"
	GetApplicationProcessesRequest                              = "GetApplicationProcesses"
	GetApplicationProcessRequest                                = "GetApplicationProcess"
	GetApplicationsRequest                                      = "GetApplications"
	GetApplicationTasksRequest                                  = "GetApplicationTasks"
	GetBuildRequest                                             = "GetBuild"
	GetDeploymentRequest                                        = "GetDeployment"
	GetDropletRequest                                           = "GetDroplet"
	GetDropletsRequest                                          = "GetDroplets"
	GetIsolationSegmentOrganizationsRequest                     = "GetIsolationSegmentOrganizations"
	GetIsolationSegmentRequest                                  = "GetIsolationSegment"
	GetIsolationSegmentsRequest                                 = "GetIsolationSegments"
	GetOrganizationRelationshipDefaultIsolationSegmentRequest   = "GetOrganizationRelationshipDefaultIsolationSegment"
	GetOrganizationsRequest                                     = "GetOrganizations"
	GetPackageRequest                                           = "GetPackage"
	GetPackagesRequest                                          = "GetPackages"
	GetProcessStatsRequest                                      = "GetProcessStats"
	GetServiceInstancesRequest                                  = "GetServiceInstances"
	GetSpaceRelationshipIsolationSegmentRequest                 = "GetSpaceRelationshipIsolationSegment"
	GetSpacesRequest                                            = "GetSpaces"
	PatchApplicationCurrentDropletRequest                       = "PatchApplicationCurrentDroplet"
	PatchApplicationEnvironmentVariablesRequest                 = "PatchApplicationEnvironmentVariables"
	PatchApplicationRequest                                     = "PatchApplication"
	PatchOrganizationRelationshipDefaultIsolationSegmentRequest = "PatchOrganizationRelationshipDefaultIsolationSegment"
	PatchProcessRequest                                         = "PatchProcess"
	PatchSpaceRelationshipIsolationSegmentRequest               = "PatchSpaceRelationshipIsolationSegment"
	PostApplicationActionApplyManifest                          = "PostApplicationActionApplyM"
	PostApplicationActionRestartRequest                         = "PostApplicationActionRestart"
	PostApplicationActionStartRequest                           = "PostApplicationActionStart"
	PostApplicationActionStopRequest                            = "PostApplicationActionStop"
	PostApplicationDeploymentRequest                            = "PostApplicationDeployment"
	PostApplicationProcessActionScaleRequest                    = "PostApplicationProcessActionScale"
	PostApplicationRequest                                      = "PostApplication"
	PostApplicationTasksRequest                                 = "PostApplicationTasks"
	PostBuildRequest                                            = "PostBuild"
	PostIsolationSegmentRelationshipOrganizationsRequest        = "PostIsolationSegmentRelationshipOrganizations"
	PostIsolationSegmentsRequest                                = "PostIsolationSegments"
	PostPackageRequest                                          = "PostPackage"
	PostServiceInstanceRelationshipsSharedSpacesRequest         = "PostServiceInstanceRelationshipsSharedSpaces"
	PutTaskCancelRequest                                        = "PutTaskCancel"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Resource: AppsResource, Path: "/", Method: http.MethodGet, Name: GetApplicationsRequest},
	{Resource: AppsResource, Path: "/", Method: http.MethodPost, Name: PostApplicationRequest},
	{Resource: AppsResource, Path: "/:app_guid", Method: http.MethodDelete, Name: DeleteApplicationRequest},
	{Resource: AppsResource, Path: "/:app_guid", Method: http.MethodPatch, Name: PatchApplicationRequest},
	{Resource: AppsResource, Path: "/:app_guid/actions/apply_manifest", Method: http.MethodPost, Name: PostApplicationActionApplyManifest},
	{Resource: AppsResource, Path: "/:app_guid/actions/restart", Method: http.MethodPost, Name: PostApplicationActionRestartRequest},
	{Resource: AppsResource, Path: "/:app_guid/actions/start", Method: http.MethodPost, Name: PostApplicationActionStartRequest},
	{Resource: AppsResource, Path: "/:app_guid/actions/stop", Method: http.MethodPost, Name: PostApplicationActionStopRequest},
	{Resource: AppsResource, Path: "/:app_guid/droplets/current", Method: http.MethodGet, Name: GetApplicationDropletCurrentRequest},
	{Resource: AppsResource, Path: "/:app_guid/env", Method: http.MethodGet, Name: GetApplicationEnvRequest},
	{Resource: AppsResource, Path: "/:app_guid/environment_variables", Method: http.MethodPatch, Name: PatchApplicationEnvironmentVariablesRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes", Method: http.MethodGet, Name: GetApplicationProcessesRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes/:type", Method: http.MethodGet, Name: GetApplicationProcessRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes/:type/actions/scale", Method: http.MethodPost, Name: PostApplicationProcessActionScaleRequest},
	{Resource: AppsResource, Path: "/:app_guid/processes/:type/instances/:index", Method: http.MethodDelete, Name: DeleteApplicationProcessInstanceRequest},
	{Resource: AppsResource, Path: "/:app_guid/relationships/current_droplet", Method: http.MethodPatch, Name: PatchApplicationCurrentDropletRequest},
	{Resource: AppsResource, Path: "/:app_guid/tasks", Method: http.MethodGet, Name: GetApplicationTasksRequest},
	{Resource: AppsResource, Path: "/:app_guid/tasks", Method: http.MethodPost, Name: PostApplicationTasksRequest},
	{Resource: BuildsResource, Path: "/", Method: http.MethodPost, Name: PostBuildRequest},
	{Resource: BuildsResource, Path: "/:build_guid", Method: http.MethodGet, Name: GetBuildRequest},
	{Resource: DeploymentsResource, Path: "/:deployment_guid", Method: http.MethodGet, Name: GetDeploymentRequest},
	{Resource: DeploymentsResource, Path: "/", Method: http.MethodPost, Name: PostApplicationDeploymentRequest},
	{Resource: DropletsResource, Path: "/", Method: http.MethodGet, Name: GetDropletsRequest},
	{Resource: DropletsResource, Path: "/:droplet_guid", Method: http.MethodGet, Name: GetDropletRequest},
	{Resource: IsolationSegmentsResource, Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest},
	{Resource: IsolationSegmentsResource, Path: "/", Method: http.MethodPost, Name: PostIsolationSegmentsRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid", Method: http.MethodGet, Name: GetIsolationSegmentRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid/organizations", Method: http.MethodGet, Name: GetIsolationSegmentOrganizationsRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid/relationships/organizations", Method: http.MethodPost, Name: PostIsolationSegmentRelationshipOrganizationsRequest},
	{Resource: IsolationSegmentsResource, Path: "/:isolation_segment_guid/relationships/organizations/:organization_guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRelationshipOrganizationRequest},
	{Resource: OrgsResource, Path: "/", Method: http.MethodGet, Name: GetOrganizationsRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/relationships/default_isolation_segment", Method: http.MethodGet, Name: GetOrganizationRelationshipDefaultIsolationSegmentRequest},
	{Resource: OrgsResource, Path: "/:organization_guid/relationships/default_isolation_segment", Method: http.MethodPatch, Name: PatchOrganizationRelationshipDefaultIsolationSegmentRequest},
	{Resource: PackagesResource, Path: "/", Method: http.MethodGet, Name: GetPackagesRequest},
	{Resource: PackagesResource, Path: "/", Method: http.MethodPost, Name: PostPackageRequest},
	{Resource: PackagesResource, Path: "/:package_guid", Method: http.MethodGet, Name: GetPackageRequest},
	{Resource: ProcessesResource, Path: "/:process_guid", Method: http.MethodPatch, Name: PatchProcessRequest},
	{Resource: ProcessesResource, Path: "/:process_guid/stats", Method: http.MethodGet, Name: GetProcessStatsRequest},
	{Resource: ServiceInstancesResource, Path: "/", Method: http.MethodGet, Name: GetServiceInstancesRequest},
	{Resource: ServiceInstancesResource, Path: "/:service_instance_guid/relationships/shared_spaces", Method: http.MethodPost, Name: PostServiceInstanceRelationshipsSharedSpacesRequest},
	{Resource: ServiceInstancesResource, Path: "/:service_instance_guid/relationships/shared_spaces/:space_guid", Method: http.MethodDelete, Name: DeleteServiceInstanceRelationshipsSharedSpaceRequest},
	{Resource: SpacesResource, Path: "/", Method: http.MethodGet, Name: GetSpacesRequest},
	{Resource: SpacesResource, Path: "/:space_guid/relationships/isolation_segment", Method: http.MethodGet, Name: GetSpaceRelationshipIsolationSegmentRequest},
	{Resource: SpacesResource, Path: "/:space_guid/relationships/isolation_segment", Method: http.MethodPatch, Name: PatchSpaceRelationshipIsolationSegmentRequest},
	{Resource: TasksResource, Path: "/:task_guid/cancel", Method: http.MethodPut, Name: PutTaskCancelRequest},
}
