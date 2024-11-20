package ccv3

import (
	"fmt"

	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) UpdateResourceMetadata(resource string, resourceGUID string, metadata resources.Metadata) (JobURL, Warnings, error) {
	var params RequestParams
	requestMetadata := resources.ResourceMetadata{Metadata: &metadata}

	switch resource {
	case "app":
		params = RequestParams{
			RequestName: ccv3internal.PatchApplicationRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"app_guid": resourceGUID},
		}
	case "buildpack":
		params = RequestParams{
			RequestName: ccv3internal.PatchBuildpackRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"buildpack_guid": resourceGUID},
		}
	case "domain":
		params = RequestParams{
			RequestName: ccv3internal.PatchDomainRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"domain_guid": resourceGUID},
		}
	case "org":
		params = RequestParams{
			RequestName: ccv3internal.PatchOrganizationRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"organization_guid": resourceGUID},
		}
	case "route":
		params = RequestParams{
			RequestName: ccv3internal.PatchRouteRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"route_guid": resourceGUID},
		}
	case "service-broker":
		params = RequestParams{
			RequestName: ccv3internal.PatchServiceBrokerRequest,
			URIParams:   internal.Params{"service_broker_guid": resourceGUID},
			RequestBody: resources.ResourceMetadata{Metadata: &metadata},
		}
	case "service-instance":
		params = RequestParams{
			RequestName: ccv3internal.PatchServiceInstanceRequest,
			URIParams:   internal.Params{"service_instance_guid": resourceGUID},
			RequestBody: resources.ResourceMetadata{Metadata: &metadata},
		}
	case "service-offering":
		params = RequestParams{
			RequestName: ccv3internal.PatchServiceOfferingRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"service_offering_guid": resourceGUID},
		}
	case "service-plan":
		params = RequestParams{
			RequestName: ccv3internal.PatchServicePlanRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"service_plan_guid": resourceGUID},
		}
	case "space":
		params = RequestParams{
			RequestName: ccv3internal.PatchSpaceRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"space_guid": resourceGUID},
		}
	case "stack":
		params = RequestParams{
			RequestName: ccv3internal.PatchStackRequest,
			RequestBody: requestMetadata,
			URIParams:   internal.Params{"stack_guid": resourceGUID},
		}
	default:
		return "", nil, fmt.Errorf("unknown resource type (%s) requested", resource)
	}

	return client.MakeRequest(params)
}
