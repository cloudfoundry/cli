package ccv3

import (
	"net/http"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
    "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
)

type Info struct {
	Name          string `json:"name"`
	Build         string `json:"build"`
	OSBAPIVersion string `json:"osbapi_version"`
}

// GetRoot returns the /v3/info response
func (client *Client) GetInfo() (Info, Warnings, error) {
	var responseBody Info

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.Info,
		ResponseBody: &responseBody,
	})

	unknownSourceErr, ok := err.(ccerror.UnknownHTTPSourceError)
	if ok && unknownSourceErr.StatusCode == http.StatusNotFound {
		return Info{}, nil, ccerror.APINotFoundError{URL: client.CloudControllerURL}
	}

	return responseBody, warnings, err
}
