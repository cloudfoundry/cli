package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// RootResponse represents a GET response from the '/' endpoint of the cloud
// controller API.
type RootResponse struct {
	Links RootLinks `json:"links"`
}

func (r RootResponse) ccV3Href() string {
	return r.Links.CCV3.HREF
}

// RootLinks represents the links in RootResponse.
type RootLinks struct {
	CCV3 APILink `json:"cloud_controller_v3"`
}

// APIInformation represents the information returned back from /v3.
type APIInformation struct {
	Links APILinks `json:"links"`
}

// Return the HREF for the UAA.
func (a APIInformation) UAA() string {
	return a.Links.UAA.HREF
}

// APILinks represents the links in APIInformation.
type APILinks struct {
	UAA APILink `json:"uaa"`
}

// APILink represents a generic link from a response object.
type APILink struct {
	HREF string `json:"href"`
}

// Info returns back endpoint and API information from /v3.
func (client *Client) Info() (APIInformation, Warnings, error) {
	rootResponse, warnings, err := client.rootResponse()
	if err != nil {
		return APIInformation{}, warnings, err
	}

	request, err := newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URI:    rootResponse.ccV3Href(),
	})
	if err != nil {
		return APIInformation{}, warnings, err
	}

	var info APIInformation
	response := cloudcontroller.Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	warnings = append(warnings, response.Warnings...)

	if err != nil {
		return APIInformation{}, warnings, err
	}

	return info, warnings, nil
}

// rootResponse return the CC API root document.
func (client *Client) rootResponse() (RootResponse, Warnings, error) {
	request, err := newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URI:    client.cloudControllerURL,
	})
	if err != nil {
		return RootResponse{}, nil, err
	}

	var rootResponse RootResponse
	response := cloudcontroller.Response{
		Result: &rootResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return RootResponse{}, response.Warnings, err
	}

	return rootResponse, response.Warnings, nil
}
