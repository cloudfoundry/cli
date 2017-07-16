package internal

import "net/http"

// Naming convention:
//
// Method + non-parameter parts of the path
//
// If the request returns a single entity by GUID, use the singular (for example
// /v2/organizations/:organization_guid is GetOrganization).
//
// The const name should always be the const value + Request.
const (
	DeleteIsolationSegmentRelationshipOrganizationRequest = "DeleteIsolationSegmentRelationshipOrganization"
	DeleteIsolationSegmentRequest                         = "DeleteIsolationSegment"
	GetAppsRequest                                        = "GetApps"
	GetAppProcessesRequest                                = "GetAppProcesses"
	GetAppTasksRequest                                    = "GetAppTasks"
	GetAppDropletCurrent                                  = "GetAppDropletCurrent"
	GetProcessInstancesRequest                            = "GetProcessInstances"
	GetBuildRequest                                       = "GetBuild"
	GetIsolationSegmentOrganizationsRequest               = "GetIsolationSegmentRelationshipOrganizations"
	GetIsolationSegmentRequest                            = "GetIsolationSegment"
	GetIsolationSegmentsRequest                           = "GetIsolationSegments"
	GetOrganizationDefaultIsolationSegmentRequest         = "GetOrganizationDefaultIsolationSegment"
	GetOrgsRequest                                        = "GetOrgs"
	GetPackageRequest                                     = "GetPackage"
	GetSpaceRelationshipIsolationSegmentRequest           = "GetSpaceRelationshipIsolationSegmentRequest"
	PatchApplicationRequest                               = "PatchApplicationRequest"
	PatchApplicationCurrentDropletRequest                 = "PatchApplicationCurrentDroplet"
	PatchOrganizationDefaultIsolationSegmentRequest       = "PatchOrganizationDefaultIsolationSegmentRequest"
	PatchSpaceRelationshipIsolationSegmentRequest         = "PatchSpaceRelationshipIsolationSegmentRequest"
	PostApplicationRequest                                = "PostApplicationRequest"
	PostAppTasksRequest                                   = "PostAppTasks"
	PostBuildRequest                                      = "PostBuild"
	PostIsolationSegmentRelationshipOrganizationsRequest  = "PostIsolationSegmentRelationshipOrganizations"
	PostIsolationSegmentsRequest                          = "PostIsolationSegments"
	PostPackageRequest                                    = "PostPackageRequest"
	PutTaskCancelRequest                                  = "PutTaskCancelRequest"
	PutApplicationStartRequest                            = "PutApplicationStart"
	PutApplicationStopRequest                             = "PutApplicationStop"
)

const (
	AppsResource              = "apps"
	BuildsResource            = "builds"
	IsolationSegmentsResource = "isolation_segments"
	OrgsResource              = "organizations"
	PackagesResource          = "packages"
	ProcessesResource         = "processes"
	SpaceResource             = "spaces"
	TasksResource             = "tasks"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/", Method: http.MethodGet, Name: GetAppsRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodGet, Name: GetOrgsRequest, Resource: OrgsResource},
	{Path: "/", Method: http.MethodPost, Name: PostApplicationRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodPost, Name: PostBuildRequest, Resource: BuildsResource},
	{Path: "/", Method: http.MethodPost, Name: PostIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodPost, Name: PostPackageRequest, Resource: PackagesResource},
	{Path: "/:guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid", Method: http.MethodGet, Name: GetBuildRequest, Resource: BuildsResource},
	{Path: "/:guid", Method: http.MethodGet, Name: GetIsolationSegmentRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid", Method: http.MethodGet, Name: GetPackageRequest, Resource: PackagesResource},
	{Path: "/:guid", Method: http.MethodPatch, Name: PatchApplicationRequest, Resource: AppsResource},
	{Path: "/:guid/cancel", Method: http.MethodPut, Name: PutTaskCancelRequest, Resource: TasksResource},
	{Path: "/:guid/droplets/current", Method: http.MethodGet, Name: GetAppDropletCurrent, Resource: AppsResource},
	{Path: "/:guid/organizations", Method: http.MethodGet, Name: GetIsolationSegmentOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/processes", Method: http.MethodGet, Name: GetAppProcessesRequest, Resource: AppsResource},
	{Path: "/:guid/stats", Method: http.MethodGet, Name: GetProcessInstancesRequest, Resource: ProcessesResource},
	{Path: "/:guid/relationships/default_isolation_segment", Method: http.MethodGet, Name: GetOrganizationDefaultIsolationSegmentRequest, Resource: OrgsResource},
	{Path: "/:guid/relationships/default_isolation_segment", Method: http.MethodPatch, Name: PatchOrganizationDefaultIsolationSegmentRequest, Resource: OrgsResource},
	{Path: "/:guid/relationships/isolation_segment", Method: http.MethodGet, Name: GetSpaceRelationshipIsolationSegmentRequest, Resource: SpaceResource},
	{Path: "/:guid/relationships/current_droplet", Method: http.MethodPatch, Name: PatchApplicationCurrentDropletRequest, Resource: AppsResource},
	{Path: "/:guid/relationships/isolation_segment", Method: http.MethodPatch, Name: PatchSpaceRelationshipIsolationSegmentRequest, Resource: SpaceResource},
	{Path: "/:guid/relationships/organizations", Method: http.MethodPost, Name: PostIsolationSegmentRelationshipOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/relationships/organizations/:org_guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRelationshipOrganizationRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/start", Method: http.MethodPut, Name: PutApplicationStartRequest, Resource: AppsResource},
	{Path: "/:guid/stop", Method: http.MethodPut, Name: PutApplicationStopRequest, Resource: AppsResource},
	{Path: "/:guid/tasks", Method: http.MethodGet, Name: GetAppTasksRequest, Resource: AppsResource},
	{Path: "/:guid/tasks", Method: http.MethodPost, Name: PostAppTasksRequest, Resource: AppsResource},
}
