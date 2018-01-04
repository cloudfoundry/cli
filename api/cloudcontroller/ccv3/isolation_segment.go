package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// IsolationSegment represents a Cloud Controller Isolation Segment.
type IsolationSegment struct {
	Name string `json:"name"`
	GUID string `json:"guid,omitempty"`
}

// CreateIsolationSegment will create an Isolation Segment on the Cloud
// Controller. Note: This will not validate that the placement tag exists in
// the diego cluster.
func (client *Client) CreateIsolationSegment(isolationSegment IsolationSegment) (IsolationSegment, Warnings, error) {
	body, err := json.Marshal(isolationSegment)
	if err != nil {
		return IsolationSegment{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostIsolationSegmentsRequest,
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return IsolationSegment{}, nil, err
	}

	var responseIsolationSegment IsolationSegment
	response := cloudcontroller.Response{
		Result: &responseIsolationSegment,
	}

	err = client.connection.Make(request, &response)
	return responseIsolationSegment, response.Warnings, err
}

// GetIsolationSegments lists isolation segments with optional filters.
func (client *Client) GetIsolationSegments(query ...Query) ([]IsolationSegment, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetIsolationSegmentsRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullIsolationSegmentsList []IsolationSegment
	warnings, err := client.paginate(request, IsolationSegment{}, func(item interface{}) error {
		if isolationSegment, ok := item.(IsolationSegment); ok {
			fullIsolationSegmentsList = append(fullIsolationSegmentsList, isolationSegment)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   IsolationSegment{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullIsolationSegmentsList, warnings, err
}

// GetIsolationSegment returns back the requested isolation segment that
// matches the GUID.
func (client *Client) GetIsolationSegment(guid string) (IsolationSegment, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetIsolationSegmentRequest,
		URIParams:   map[string]string{"isolation_segment_guid": guid},
	})
	if err != nil {
		return IsolationSegment{}, nil, err
	}
	var isolationSegment IsolationSegment
	response := cloudcontroller.Response{
		Result: &isolationSegment,
	}

	err = client.connection.Make(request, &response)
	if err != nil {
		return IsolationSegment{}, response.Warnings, err
	}

	return isolationSegment, response.Warnings, nil
}

// DeleteIsolationSegment removes an isolation segment from the cloud
// controller. Note: This will only remove it from the cloud controller
// database. It will not remove it from diego.
func (client *Client) DeleteIsolationSegment(guid string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteIsolationSegmentRequest,
		URIParams:   map[string]string{"isolation_segment_guid": guid},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)
	return response.Warnings, err
}
