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
	var (
		params           requestParams
		err              error
		responseMetadata ResourceMetadata
		warnings         Warnings
		requestMetadata  ResourceMetadata
	)

	requestMetadata = ResourceMetadata{Metadata: &metadata}

	switch resource {
	case "app":
		params = requestParams{
			RequestName:  internal.PatchApplicationRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"app_guid": resourceGUID},
		}
	case "buildpack":
		params = requestParams{
			RequestName:  internal.PatchBuildpackRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"buildpack_guid": resourceGUID},
		}
	case "domain":
		params = requestParams{
			RequestName:  internal.PatchDomainRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"domain_guid": resourceGUID},
		}
	case "org":
		params = requestParams{
			RequestName:  internal.PatchOrganizationRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"organization_guid": resourceGUID},
		}
	case "route":
		params = requestParams{
			RequestName:  internal.PatchRouteRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"route_guid": resourceGUID},
		}
	case "service-offering":
		params = requestParams{
			RequestName:  internal.PatchServiceOfferingRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"service_offering_guid": resourceGUID},
		}
	case "service-plan":
		params = requestParams{
			RequestName:  internal.PatchServicePlanRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"service_plan_guid": resourceGUID},
		}
	case "space":
		params = requestParams{
			RequestName:  internal.PatchSpaceRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"space_guid": resourceGUID},
		}
	case "stack":
		params = requestParams{
			RequestName:  internal.PatchStackRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"stack_guid": resourceGUID},
		}
	default:
		return ResourceMetadata{}, nil, fmt.Errorf("unknown resource type (%s) requested", resource)
	}

	_, warnings, err = client.makeRequest(params)

	return responseMetadata, warnings, err
}

func (client *Client) UpdateResourceMetadataAsync(resource string, resourceGUID string, metadata Metadata) (JobURL, Warnings, error) {
	metadataBytes, err := json.Marshal(ResourceMetadata{Metadata: &metadata})
	if err != nil {
		return "", nil, err
	}

	var request *cloudcontroller.Request

	switch resource {
	case "service-broker":
		request, err = client.newHTTPRequest(requestOptions{
			RequestName: internal.PatchServiceBrokerRequest,
			Body:        bytes.NewReader(metadataBytes),
			URIParams:   internal.Params{"service_broker_guid": resourceGUID},
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
