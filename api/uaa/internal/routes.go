package internal

import "github.com/tedsuo/rata"

const (
	RefreshTokenRequest = "RefreshToken"
)

var Routes = rata.Routes{
	{Path: "/oauth/token", Method: "POST", Name: RefreshTokenRequest},
}
