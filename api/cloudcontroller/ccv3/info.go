package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// APILink represents a generic link from a response object.
type APILink struct {
	// HREF is the fully qualified URL for the link.
	HREF string `json:"href"`
}

// RootResponse represents a GET response from the '/' endpoint of the cloud
// controller API.
type RootResponse struct {
	// Links is a list of top level Cloud Controller APIs.
	Links struct {
		// CCV3 is the link to the Cloud Controller V3 API
		CCV3 APILink `json:"cloud_controller_v3"`

		// UAA is the link to the UAA API
		UAA APILink `json:"uaa"`
	} `json:"links"`
}

// Return the HREF for the UAA.
func (root RootResponse) UAA() string {
	return root.Links.UAA.HREF
}

func (r RootResponse) ccV3Link() string {
	return r.Links.CCV3.HREF
}

// ResourceLinks represents the information returned back from /v3.
type ResourceLinks struct {
	// Links is a list of top level Cloud Controller resources endpoints.
	Links struct {
	} `json:"links"`
}

// Info returns back endpoint and API information from /v3.
func (client *Client) Info() (RootResponse, ResourceLinks, Warnings, error) {
	rootResponse, warnings, err := client.rootResponse()
	if err != nil {
		return RootResponse{}, ResourceLinks{}, warnings, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URL:    rootResponse.ccV3Link(),
	})
	if err != nil {
		return RootResponse{}, ResourceLinks{}, warnings, err
	}

	var info ResourceLinks
	response := cloudcontroller.Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	warnings = append(warnings, response.Warnings...)

	if err != nil {
		return RootResponse{}, ResourceLinks{}, warnings, err
	}

	return rootResponse, info, warnings, nil
}

// rootResponse return the CC API root document.
func (client *Client) rootResponse() (RootResponse, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URL:    client.cloudControllerURL,
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

// UAA returns back the location of the UAA Endpoint
func (client *Client) UAA() string {
	return client.uaaLink
}
