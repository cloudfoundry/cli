package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

const (
	PostUserRequest       = "PostUser"
	PostOAuthTokenRequest = "PostOAuthToken"
)

// Routes is a list of routes used by the rata library to construct request
// URLs.
var Routes = rata.Routes{
	{Path: "/Users", Method: http.MethodPost, Name: PostUserRequest},
	{Path: "/oauth/token", Method: http.MethodPost, Name: PostOAuthTokenRequest},
}
