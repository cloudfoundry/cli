package internal

import (
	"net/http"
)

const (
	GetSSHPasscodeRequest = "GetSSHPasscode"
	PostOAuthTokenRequest = "PostOAuthToken"
	PostUserRequest       = "PostUser"
	ListUsersRequest      = "ListUsers"
	DeleteUserRequest     = "DeleteUser"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/Users", Method: http.MethodPost, Name: PostUserRequest, Resource: UAAResource},
	{Path: "/Users", Method: http.MethodGet, Name: ListUsersRequest, Resource: UAAResource},
	{Path: "/Users/:user_guid", Method: http.MethodDelete, Name: DeleteUserRequest, Resource: UAAResource},
	{Path: "/oauth/authorize", Method: http.MethodGet, Name: GetSSHPasscodeRequest, Resource: UAAResource},
	{Path: "/oauth/token", Method: http.MethodPost, Name: PostOAuthTokenRequest, Resource: AuthorizationResource},
}
