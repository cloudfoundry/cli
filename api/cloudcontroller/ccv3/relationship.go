package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

// DeleteIsolationSegmentOrganization will delete the relationship between
// the isolation segment and the organization provided.
func (client *Client) DeleteIsolationSegmentOrganization(isolationSegmentGUID string, orgGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteIsolationSegmentRelationshipOrganizationRequest,
		URIParams:   internal.Params{"isolation_segment_guid": isolationSegmentGUID, "organization_guid": orgGUID},
	})

	return warnings, err
}

// DeleteServiceInstanceRelationshipsSharedSpace will delete the sharing relationship
// between the service instance and the shared-to space provided.
func (client *Client) DeleteServiceInstanceRelationshipsSharedSpace(serviceInstanceGUID string, spaceGUID string) (Warnings, error) {
	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.DeleteServiceInstanceRelationshipsSharedSpaceRequest,
		URIParams:   internal.Params{"service_instance_guid": serviceInstanceGUID, "space_guid": spaceGUID},
	})

	return warnings, err
}

// GetOrganizationDefaultIsolationSegment returns the relationship between an
// organization and it's default isolation segment.
func (client *Client) GetOrganizationDefaultIsolationSegment(orgGUID string) (resources.Relationship, Warnings, error) {
	var responseBody resources.Relationship

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetOrganizationRelationshipDefaultIsolationSegmentRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetSpaceIsolationSegment returns the relationship between a space and it's
// isolation segment.
func (client *Client) GetSpaceIsolationSegment(spaceGUID string) (resources.Relationship, Warnings, error) {
	var responseBody resources.Relationship

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetSpaceRelationshipIsolationSegmentRequest,
		URIParams:    internal.Params{"space_guid": spaceGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// SetApplicationDroplet sets the specified droplet on the given application.
func (client *Client) SetApplicationDroplet(appGUID string, dropletGUID string) (resources.Relationship, Warnings, error) {
	var responseBody resources.Relationship

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchApplicationCurrentDropletRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		RequestBody:  resources.Relationship{GUID: dropletGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateOrganizationDefaultIsolationSegmentRelationship sets the default isolation segment
// for an organization on the controller.
// If isoSegGuid is empty it will reset the default isolation segment.
func (client *Client) UpdateOrganizationDefaultIsolationSegmentRelationship(orgGUID string, isoSegGUID string) (resources.Relationship, Warnings, error) {
	var responseBody resources.Relationship

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchOrganizationRelationshipDefaultIsolationSegmentRequest,
		URIParams:    internal.Params{"organization_guid": orgGUID},
		RequestBody:  resources.Relationship{GUID: isoSegGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// UpdateSpaceIsolationSegmentRelationship assigns an isolation segment to a space and
// returns the relationship.
func (client *Client) UpdateSpaceIsolationSegmentRelationship(spaceGUID string, isolationSegmentGUID string) (resources.Relationship, Warnings, error) {
	var responseBody resources.Relationship

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PatchSpaceRelationshipIsolationSegmentRequest,
		URIParams:    internal.Params{"space_guid": spaceGUID},
		RequestBody:  resources.Relationship{GUID: isolationSegmentGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
