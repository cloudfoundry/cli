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
	GetIsolationSegmentRequest                            = "GetIsolationSegment"
	GetIsolationSegmentsRequest                           = "GetIsolationSegments"
	GetOrgsRequest                                        = "GetOrgs"
	GetPackageRequest                                     = "GetPackage"
	GetSpaceRelationshipIsolationSegmentRequest           = "GetSpaceRelationshipIsolationSegmentRequest"
	PatchSpaceRelationshipIsolationSegmentRequest         = "PatchSpaceRelationshipIsolationSegmentRequest"
	PostApplicationRequest                                = "PostApplicationRequest"
	PostAppTasksRequest                                   = "PostAppTasks"
	PostIsolationSegmentRelationshipOrganizationsRequest  = "PostIsolationSegmentRelationshipOrganizations"
	PostIsolationSegmentsRequest                          = "PostIsolationSegments"
	PostPackageRequest                                    = "PostPackageRequest"
)

const (
	AppsResource              = "apps"
	IsolationSegmentsResource = "isolation_segments"
	OrgsResource              = "organizations"
	PackagesResource          = "packages"
	SpaceResource             = "spaces"
	TasksResource             = "tasks"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/:guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid", Method: http.MethodGet, Name: GetIsolationSegmentRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid", Method: http.MethodGet, Name: GetPackageRequest, Resource: PackagesResource},
	{Path: "/:guid/organizations", Method: http.MethodGet, Name: GetIsolationSegmentOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/relationships/isolation_segment", Method: http.MethodGet, Name: GetSpaceRelationshipIsolationSegmentRequest, Resource: SpaceResource},
	{Path: "/:guid/relationships/isolation_segment", Method: http.MethodPatch, Name: PatchSpaceRelationshipIsolationSegmentRequest, Resource: SpaceResource},
	{Path: "/:guid/relationships/organizations", Method: http.MethodPost, Name: PostIsolationSegmentRelationshipOrganizationsRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/relationships/organizations/:org_guid", Method: http.MethodDelete, Name: DeleteIsolationSegmentRelationshipOrganizationRequest, Resource: IsolationSegmentsResource},
	{Path: "/:guid/tasks", Method: http.MethodGet, Name: GetAppTasksRequest, Resource: AppsResource},
	{Path: "/:guid/tasks", Method: http.MethodPost, Name: PostAppTasksRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetAppsRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodGet, Name: GetIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodGet, Name: GetOrgsRequest, Resource: OrgsResource},
	{Path: "/", Method: http.MethodPost, Name: PostApplicationRequest, Resource: AppsResource},
	{Path: "/", Method: http.MethodPost, Name: PostIsolationSegmentsRequest, Resource: IsolationSegmentsResource},
	{Path: "/", Method: http.MethodPost, Name: PostPackageRequest, Resource: PackagesResource},
}
