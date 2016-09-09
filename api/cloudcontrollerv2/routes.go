package cloudcontrollerv2

import "github.com/tedsuo/rata"

const (
	InfoRequest = "Info"
)

var Routes = rata.Routes{
	{Path: "/v2/info", Method: "GET", Name: InfoRequest},
}
