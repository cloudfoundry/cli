package internal

import "net/http"

const (
	DeleteIsolationSegmentRequest = "DeleteIsolationSegment"
	GetAppTasksRequest            = "AppTasks"
	GetAppsRequest                = "Apps"
	GetIsolationSegmentsRequest   = "GetIsolationSegments"
	NewAppTaskRequest             = "NewAppTask"
	NewIsolationSegmentRequest    = "NewIsolationSegment"
)

const (
	AppsResource              = "apps"
	TasksResource             = "tasks"
	IsolationSegmentsResource = "isolation_segments"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/:guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/tasks", Method: http.MethodGet, Name: GetAppTasksRequest, Resource: AppsResource},
	{Path: "/:guid/tasks", Method: http.MethodPost, Name: NewAppTaskRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetAppsRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodPost, Name: NewIsolationSegmentRequest, Resource: IsolationSegmentsResource},
}
