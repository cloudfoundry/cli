package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// IsolationSegment represents a Cloud Controller Isolation Segment
type IsolationSegment struct {
	Name string `json:"name"`
	GUID string `json:"guid,omitempty"`
}

// GetApplications lists applications with optional filters.
func (client *Client) CreateIsolationSegment(name string) (IsolationSegment, Warnings, error) {
	body, err := json.Marshal(IsolationSegment{Name: name})
	if err != nil {
		return IsolationSegment{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.NewIsolationSegmentRequest,
		Body:        bytes.NewBuffer(body),
	})
	if err != nil {
		return IsolationSegment{}, nil, err
	}

	var isolationSegment IsolationSegment
	response := cloudcontroller.Response{
		Result: &isolationSegment,
	}

	err = client.connection.Make(request, &response)
	return isolationSegment, response.Warnings, err
}
