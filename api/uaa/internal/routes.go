package internal

import (
	"net/http"
)

const (
	GetSSHPasscodeRequest = "GetSSHPasscode"
	PostOAuthTokenRequest = "PostOAuthToken"
	PostUserRequest       = "PostUser"
)

const (
	AuthorizationResource = "authorization_endpoint"
	UAAResource           = "uaa"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/Users", Method: http.MethodPost, Name: PostUserRequest, Resource: UAAResource},
	{Path: "/oauth/authorize", Method: http.MethodGet, Name: GetSSHPasscodeRequest, Resource: UAAResource},
	{Path: "/oauth/token", Method: http.MethodPost, Name: PostOAuthTokenRequest, Resource: AuthorizationResource},
}
