package ccv3

import (
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/internal"
	"code.cloudfoundry.org/cli/v8/resources"
)

func (client *Client) GetProcessSidecars(processGuid string) ([]resources.Sidecar, Warnings, error) {
	var sidecars []resources.Sidecar

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetProcessSidecarsRequest,
		URIParams:    internal.Params{"process_guid": processGuid},
		ResponseBody: resources.Sidecar{},
		AppendToList: func(item interface{}) error {
			sidecars = append(sidecars, item.(resources.Sidecar))
			return nil
		},
	})

	return sidecars, warnings, err
}
