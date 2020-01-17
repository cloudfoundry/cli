package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// IsolationSegment represents a Cloud Controller Isolation Segment.
type IsolationSegment struct {
	//GUID is the unique ID of the isolation segment.
	GUID string `json:"guid,omitempty"`
	//Name is the name of the isolation segment.
	Name string `json:"name"`
}

// CreateIsolationSegment will create an Isolation Segment on the Cloud
// Controller. Note: This will not validate that the placement tag exists in
// the diego cluster.
func (client *Client) CreateIsolationSegment(isolationSegment IsolationSegment) (IsolationSegment, Warnings, error) {
	var responseBody IsolationSegment

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.PostIsolationSegmentsRequest,
		RequestBody:  isolationSegment,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// DeleteIsolationSegment removes an isolation segment from the cloud
// controller. Note: This will only remove it from the cloud controller
// database. It will not remove it from diego.
func (client *Client) DeleteIsolationSegment(guid string) (Warnings, error) {
	_, warnings, err := client.makeRequest(requestParams{
		RequestName: internal.DeleteIsolationSegmentRequest,
		URIParams:   internal.Params{"isolation_segment_guid": guid},
	})

	return warnings, err
}

// GetIsolationSegment returns back the requested isolation segment that
// matches the GUID.
func (client *Client) GetIsolationSegment(guid string) (IsolationSegment, Warnings, error) {
	var responseBody IsolationSegment

	_, warnings, err := client.makeRequest(requestParams{
		RequestName:  internal.GetIsolationSegmentRequest,
		URIParams:    internal.Params{"isolation_segment_guid": guid},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetIsolationSegments lists isolation segments with optional filters.
func (client *Client) GetIsolationSegments(query ...Query) ([]IsolationSegment, Warnings, error) {
	var isolationSegments []IsolationSegment

	resources, warnings, err := client.makeListRequest(requestParams{
		RequestName:  internal.GetIsolationSegmentsRequest,
		Query:        query,
		ResponseBody: IsolationSegment{},
	})

	for _, resource := range resources {
		if isolationSegment, ok := resource.(IsolationSegment); ok {
			isolationSegments = append(isolationSegments, isolationSegment)
		} else {
			return nil, nil, ccerror.UnknownObjectInListError{
				Expected:   IsolationSegment{},
				Unexpected: isolationSegment,
			}
		}
	}

	return isolationSegments, warnings, err
}
