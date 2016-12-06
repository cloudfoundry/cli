package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

const (
	RefreshTokenRequest = "RefreshToken"
	NewUserRequest      = "NewUser"
)

var Routes = rata.Routes{
	{Path: "/oauth/token", Method: http.MethodPost, Name: RefreshTokenRequest},
	{Path: "/Users", Method: http.MethodPost, Name: NewUserRequest},
}
