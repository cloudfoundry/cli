package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

// APILink represents a generic link from a response object.
type APILink struct {
	// HREF is the fully qualified URL for the link.
	HREF string `json:"href"`

	// Meta contains additional metadata about the API.
	Meta struct {
		// Version of the API
		Version string `json:"version"`
	} `json:"meta"`
}

// APIInfo represents a GET response from the '/' endpoint of the cloud
// controller API.
type APIInfo struct {
	// Links is a list of top level Cloud Controller APIs.
	Links struct {
		// CCV3 is the link to the Cloud Controller V3 API
		CCV3 APILink `json:"cloud_controller_v3"`

		// UAA is the link to the UAA API
		UAA APILink `json:"uaa"`
	} `json:"links"`
}

// UAA return the HREF for the UAA.
func (info APIInfo) UAA() string {
	return info.Links.UAA.HREF
}

// CloudControllerAPIVersion return the version for the CloudController.
func (info APIInfo) CloudControllerAPIVersion() string {
	return info.Links.CCV3.Meta.Version
}

func (info APIInfo) ccV3Link() string {
	return info.Links.CCV3.HREF
}

// ResourceLinks represents the information returned back from /v3.
type ResourceLinks struct {
	// Links is a list of top level Cloud Controller resources endpoints.
	Links struct {
	} `json:"links"`
}

// Info returns back endpoint and API information from /v3.
func (client *Client) Info() (APIInfo, ResourceLinks, Warnings, error) {
	rootResponse, warnings, err := client.rootResponse()
	if err != nil {
		return APIInfo{}, ResourceLinks{}, warnings, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URL:    rootResponse.ccV3Link(),
	})
	if err != nil {
		return APIInfo{}, ResourceLinks{}, warnings, err
	}

	var info ResourceLinks
	response := cloudcontroller.Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	warnings = append(warnings, response.Warnings...)

	if err != nil {
		return APIInfo{}, ResourceLinks{}, warnings, err
	}

	return rootResponse, info, warnings, nil
}

// rootResponse return the CC API root document.
func (client *Client) rootResponse() (APIInfo, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		Method: http.MethodGet,
		URL:    client.cloudControllerURL,
	})
	if err != nil {
		return APIInfo{}, nil, err
	}

	var rootResponse APIInfo
	response := cloudcontroller.Response{
		Result: &rootResponse,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		if _, ok := err.(cloudcontroller.NotFoundError); ok {
			return APIInfo{}, nil, cloudcontroller.APINotFoundError{URL: client.cloudControllerURL}
		}
		return APIInfo{}, response.Warnings, err
	}

	return rootResponse, response.Warnings, nil
}
