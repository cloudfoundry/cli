package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	// "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// APIInformation represents the information returned back from /v2/info
type APIInformation struct {
	Links APILinks `json:"links"`
}

func (a APIInformation) UAA() string {
	uaa := a.Links.UAA
	foo := APILink{}
	if uaa != foo {
		return uaa.HREF
	} else {
		return ""
	}
}

type APILinks struct {
	UAA APILink `json:"uaa"`
}

type APILink struct {
	HREF string `json:"href"`
}

// TokenEndpoint returns the Token endpoint for the targeted Cloud Controller.
// func (client *CloudControllerClient) UAAEndpoint() string {
// 	return client.UAA
// }

// Info returns back endpoint and API information from /v2/info.
func (client *CloudControllerClient) Info() (APIInformation, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URI:    "/v3/",
	})
	if err != nil {
		return APIInformation{}, nil, err
	}

	var info APIInformation
	response := cloudcontroller.Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	return info, response.Warnings, err
}
