package internal

import "net/http"

const (
	DeleteIsolationSegmentRequest                        = "DeleteIsolationSegment"
	GetAppsRequest                                       = "Apps"
	GetAppTasksRequest                                   = "AppTasks"
	GetIsolationSegmentsRequest                          = "GetIsolationSegments"
	GetOrgsRequest                                       = "Orgs"
	PostAppTasksRequest                                  = "PostAppTasks"
	PostIsolationSegmentsRequest                         = "PostIsolationSegment"
	PostIsolationSegmentRelationshipOrganizationsRequest = "NewIsolationSegmentOrganizationRelationship"
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
	{Path: "/:guid/tasks", Method: http.MethodGet, Name: GetAppTasksRequest, Resource: AppsResource},
	{Path: "/:guid/tasks", Method: http.MethodPost, Name: PostAppTasksRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetAppsRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodGet, Name: GetOrgsRequest, Resource: OrgsResource},
	{Path: "/", Method: http.MethodPost, Name: PostIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
}
