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
	GetAppTasksRequest                                    = "GetAppTasks"
	GetIsolationSegmentOrganizationsRequest               = "GetIsolationSegmentRelationshipOrganizations"
	GetIsolationSegmentsRequest                           = "GetIsolationSegments"
	GetOrgsRequest                                        = "GetOrgs"
	PatchSpaceRelationshipIsolationSegmentRequest         = "PatchSpaceRelationshipIsolationSegmentRequest"
	PostAppTasksRequest                                   = "PostAppTasks"
	PostIsolationSegmentRelationshipOrganizationsRequest  = "PostIsolationSegmentRelationshipOrganizations"
	PostIsolationSegmentsRequest                          = "PostIsolationSegments"
)

const (
	AppsResource              = "apps"
	IsolationSegmentsResource = "isolation_segments"
	OrgsResource              = "organizations"
	SpaceResource             = "spaces"
	TasksResource             = "tasks"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/:guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/organizations", Method: http.MethodGet, Name: GetIsolationSegmentOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/relationships/isolation_segment", Method: http.MethodPatch, Name: PatchSpaceRelationshipIsolationSegmentRequest, Resource: SpaceResource},
	{Path: "/:guid/relationships/organizations", Method: http.MethodPost, Name: PostIsolationSegmentRelationshipOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/relationships/organizations/:org_guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRelationshipOrganizationRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/tasks", Method: http.MethodGet, Name: GetAppTasksRequest, Resource: AppsResource},
	{Path: "/:guid/tasks", Method: http.MethodPost, Name: PostAppTasksRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetAppsRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodGet, Name: GetOrgsRequest, Resource: OrgsResource},
	{Path: "/", Method: http.MethodPost, Name: PostIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
}
