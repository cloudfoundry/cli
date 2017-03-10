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
	DeleteIsolationSegmentRequest                        = "DeleteIsolationSegment"
	GetAppsRequest                                       = "GetApps"
	GetAppTasksRequest                                   = "GetAppTasks"
	GetIsolationSegmentsRequest                          = "GetIsolationSegments"
	GetIsolationSegmentOrganizationsRequest              = "GetIsolationSegmentRelationshipOrganizations"
	GetOrgsRequest                                       = "GetOrgs"
	PostAppTasksRequest                                  = "PostAppTasks"
	PostIsolationSegmentsRequest                         = "PostIsolationSegments"
	PostIsolationSegmentRelationshipOrganizationsRequest = "PostIsolationSegmentRelationshipOrganizations"
)

const (
	AppsResource              = "apps"
	IsolationSegmentsResource = "isolation_segments"
	OrgsResource              = "organizations"
	TasksResource             = "tasks"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/:guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/relationships/organizations", Method: http.MethodPost, Name: PostIsolationSegmentRelationshipOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/organizations", Method: http.MethodGet, Name: GetIsolationSegmentOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/tasks", Method: http.MethodGet, Name: GetAppTasksRequest, Resource: AppsResource},
	{Path: "/:guid/tasks", Method: http.MethodPost, Name: PostAppTasksRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetAppsRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodGet, Name: GetOrgsRequest, Resource: OrgsResource},
	{Path: "/", Method: http.MethodPost, Name: PostIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
}
