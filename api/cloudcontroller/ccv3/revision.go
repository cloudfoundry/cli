package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/resources"
)

func (client *Client) GetRevision(revisionGUID string) (resources.Revision, Warnings, error) {
	var revision resources.Revision

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetRevisionRequest,
		URIParams:    internal.Params{"revision_guid": revisionGUID},
		ResponseBody: &revision,
	})

	return revision, warnings, err
}
