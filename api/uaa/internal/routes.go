package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

const (
	GetSSHPasscodeRequest = "GetSSHPasscode"
	PostOAuthTokenRequest = "PostOAuthToken"
	PostUserRequest       = "PostUser"
)

// Routes is a list of routes used by the rata library to construct request
// URLs.
var Routes = rata.Routes{
	{Path: "/Users", Method: http.MethodPost, Name: PostUserRequest},
	{Path: "/oauth/authorize", Method: http.MethodGet, Name: GetSSHPasscodeRequest},
	{Path: "/oauth/token", Method: http.MethodPost, Name: PostOAuthTokenRequest},
}
