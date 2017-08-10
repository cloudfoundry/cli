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
	UAAResource   = "uaa"
	LoginResource = "login"
)

// APIRoutes is a list of routes used by the router to construct request URLs.
var APIRoutes = []Route{
	{Path: "/Users", Method: http.MethodPost, Name: PostUserRequest, Resource: UAAResource},
	{Path: "/oauth/authorize", Method: http.MethodGet, Name: GetSSHPasscodeRequest, Resource: LoginResource},
	{Path: "/oauth/token", Method: http.MethodPost, Name: PostOAuthTokenRequest, Resource: LoginResource},
}
