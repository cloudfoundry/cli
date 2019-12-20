package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/types"
)

// Metadata is used for custom tagging of API resources
type Metadata struct {
	Labels map[string]types.NullString `json:"labels,omitempty"`
}

type ResourceMetadata struct {
	Metadata *Metadata `json:"metadata,omitempty"`
}

func (client *Client) UpdateResourceMetadata(resource string, resourceGUID string, metadata Metadata) (ResourceMetadata, Warnings, error) {
	metadataBytes, err := json.Marshal(ResourceMetadata{Metadata: &metadata})
	if err != nil {
		return ResourceMetadata{}, nil, err
	}

	var request *cloudcontroller.Request

	switch resource {
	case "app":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchApplicationRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"app_guid": resourceGUID},
		})
	case "buildpack":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchBuildpackRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"buildpack_guid": resourceGUID},
		})
	case "domain":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchDomainRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"domain_guid": resourceGUID},
		})
	case "org":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchOrganizationRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"organization_guid": resourceGUID},
		})
	case "route":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchRouteRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"route_guid": resourceGUID},
		})
	case "space":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchSpaceRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"space_guid": resourceGUID},
		})
	case "stack":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchStackRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"stack_guid": resourceGUID},
		})
	default:
		return ResourceMetadata{}, nil, fmt.Errorf("unknown resource type (%s) requested", resource)
	}

	if err != nil {
		return ResourceMetadata{}, nil, err
	}

	var responseMetadata ResourceMetadata
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &responseMetadata,
	}
	err = client.connection.Make(request, &response)
	return responseMetadata, response.Warnings, err
}

func (client *Client) UpdateResourceMetadataAsync(resource string, resourceGUID string, metadata Metadata) (JobURL, Warnings, error) {
	metadataBytes, err := json.Marshal(ResourceMetadata{Metadata: &metadata})
	if err != nil {
		return "", nil, err
	}

	var request *cloudcontroller.Request

	switch resource {
	case "service-broker":
		request, _ = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchServiceBrokerRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   map[string]string{"service_broker_guid": resourceGUID},
		})
	default:
		return "", nil, fmt.Errorf("unknown async resource type (%s) requested", resource)
	}

	if err != nil {
		return "", nil, err
	}

	response := cloudcontroller.Response{}
	err = client.connection.Make(request, &response)
	return JobURL(response.ResourceLocationURL), response.Warnings, err
}
