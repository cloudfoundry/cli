package ccv3

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type Process struct {
	GUID string `json:"guid"`
	Type string `json:"type"`
}

// GetApplicationProcesses lists processes for a given app
func (client *Client) GetApplicationProcesses(appGUID string) ([]Process, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetAppProcessesRequest,
		URIParams:   map[string]string{"guid": appGUID},
	})
	if err != nil {
		return nil, nil, err
	}

	var fullProcessesList []Process
	warnings, err := client.paginate(request, Process{}, func(item interface{}) error {
		if process, ok := item.(Process); ok {
			fullProcessesList = append(fullProcessesList, process)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Process{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullProcessesList, warnings, err
}
