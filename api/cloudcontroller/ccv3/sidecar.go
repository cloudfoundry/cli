package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"

	"code.cloudfoundry.org/cli/types"
)

type Sidecar struct {
	GUID    string               `json:"guid"`
	Name    string               `json:"name"`
	Command types.FilteredString `json:"command"`
}

func (client *Client) GetProcessSidecars(processGuid string) ([]Sidecar, Warnings, error) {
	request, err := client.NewHTTPRequest(requestOptions{
		RequestName: internal.GetProcessSidecarsRequest,
		URIParams:   map[string]string{"process_guid": processGuid},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullSidecarList []Sidecar
	warnings, err := client.paginate(request, Sidecar{}, func(item interface{}) error {
		if sidecar, ok := item.(Sidecar); ok {
			fullSidecarList = append(fullSidecarList, sidecar)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Sidecar{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullSidecarList, warnings, err
}
