package ccv3

import (
	ccv3internal "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/api/internal"
	"code.cloudfoundry.org/cli/resources"
)

// EntitleIsolationSegmentToOrganizations will create a link between the
// isolation segment and the list of organizations provided.
func (client *Client) EntitleIsolationSegmentToOrganizations(isolationSegmentGUID string, organizationGUIDs []string) (resources.RelationshipList, Warnings, error) {
	var responseBody resources.RelationshipList

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  ccv3internal.PostIsolationSegmentRelationshipOrganizationsRequest,
		URIParams:    internal.Params{"isolation_segment_guid": isolationSegmentGUID},
		RequestBody:  resources.RelationshipList{GUIDs: organizationGUIDs},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}
