package ccv3

import (
	"fmt"

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
		params           RequestParams
		err              error
		responseMetadata ResourceMetadata
		warnings         Warnings
		requestMetadata  ResourceMetadata
	)

	requestMetadata = ResourceMetadata{Metadata: &metadata}

	switch resource {
	case "app":
		params = RequestParams{
			RequestName:  internal.PatchApplicationRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"app_guid": resourceGUID},
		}
	case "buildpack":
		params = RequestParams{
			RequestName:  internal.PatchBuildpackRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"buildpack_guid": resourceGUID},
		}
	case "domain":
		params = RequestParams{
			RequestName:  internal.PatchDomainRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"domain_guid": resourceGUID},
		}
	case "org":
		params = RequestParams{
			RequestName:  internal.PatchOrganizationRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"organization_guid": resourceGUID},
		}
	case "route":
		params = RequestParams{
			RequestName:  internal.PatchRouteRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"route_guid": resourceGUID},
		}
	case "service-offering":
		params = RequestParams{
			RequestName:  internal.PatchServiceOfferingRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"service_offering_guid": resourceGUID},
		}
	case "service-plan":
		params = RequestParams{
			RequestName:  internal.PatchServicePlanRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"service_plan_guid": resourceGUID},
		}
	case "space":
		params = RequestParams{
			RequestName:  internal.PatchSpaceRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"space_guid": resourceGUID},
		}
	case "stack":
		params = RequestParams{
			RequestName:  internal.PatchStackRequest,
			RequestBody:  requestMetadata,
			ResponseBody: &responseMetadata,
			URIParams:    internal.Params{"stack_guid": resourceGUID},
		}
	default:
		return ResourceMetadata{}, nil, fmt.Errorf("unknown resource type (%s) requested", resource)
	}

	_, warnings, err = client.MakeRequest(params)

	return responseMetadata, warnings, err
}

func (client *Client) UpdateResourceMetadataAsync(resource string, resourceGUID string, metadata Metadata) (JobURL, Warnings, error) {
	var params RequestParams

	switch resource {
	case "service-broker":
		params = RequestParams{
			RequestName: internal.PatchServiceBrokerRequest,
			URIParams:   internal.Params{"service_broker_guid": resourceGUID},
			RequestBody: ResourceMetadata{Metadata: &metadata},
		}
	default:
		return "", nil, fmt.Errorf("unknown async resource type (%s) requested", resource)
	}

	jobURL, warnings, err := client.MakeRequest(params)

	return jobURL, warnings, err
}
