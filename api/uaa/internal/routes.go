package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

const (
	RefreshTokenRequest = "RefreshToken"
)

var Routes = rata.Routes{
	{Path: "/oauth/token", Method: http.MethodPost, Name: RefreshTokenRequest},
}
