package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"

	"code.cloudfoundry.org/cli/types"
)

type Sidecar struct {
	GUID    string               `json:"guid"`
	Name    string               `json:"name"`
	Command types.FilteredString `json:"command"`
}

func (client *Client) GetProcessSidecars(processGuid string) ([]Sidecar, Warnings, error) {
	var resources []Sidecar

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetProcessSidecarsRequest,
		URIParams:    internal.Params{"process_guid": processGuid},
		ResponseBody: Sidecar{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Sidecar))
			return nil
		},
	})

	return resources, warnings, err
}
