package ccv3

import (
	"encoding/json"
	"net/http"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
)

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
type ResourceLinks map[string]APILink

// UnmarshalJSON helps unmarshal a Cloud Controller /v3 response.
func (resources ResourceLinks) UnmarshalJSON(data []byte) error {
	var ccResourceLinks struct {
		Links map[string]APILink `json:"links"`
	}
	if err := json.Unmarshal(data, &ccResourceLinks); err != nil {
		return err
	}

	for key, val := range ccResourceLinks.Links {
		resources[key] = val
	}

	return nil
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

	info := ResourceLinks{} // Explicitly initializing
	response := cloudcontroller.Response{
		Result: &info,
	}

	err = client.connection.Make(request, &response)
	warnings = append(warnings, response.Warnings...)

	if err != nil {
		return APIInfo{}, ResourceLinks{}, warnings, err
	}

	// TODO: Remove this hack after CC adds proper IsolationSegment,
	// Organizations, and Spaces resources to /v3.
	info["isolation_segments"] = APILink{HREF: strings.Replace(info["tasks"].HREF, "tasks", "isolation_segments", 1)}
	info["organizations"] = APILink{HREF: strings.Replace(info["tasks"].HREF, "tasks", "organizations", 1)}
	info["spaces"] = APILink{HREF: strings.Replace(info["tasks"].HREF, "tasks", "spaces", 1)}

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
