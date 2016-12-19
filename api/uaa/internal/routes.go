package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

const (
	RefreshTokenRequest = "RefreshToken"
	NewUserRequest      = "NewUser"
)

// Routes is a list of routes used by the rata library to construct request
// URLs.
var Routes = rata.Routes{
	{Path: "/oauth/token", Method: http.MethodPost, Name: RefreshTokenRequest},
	{Path: "/Users", Method: http.MethodPost, Name: NewUserRequest},
}
