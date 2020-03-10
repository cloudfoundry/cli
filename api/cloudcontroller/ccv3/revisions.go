package ccv3

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"

type Revision struct {
	GUID        string
	Version     int
	Description string
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (client *Client) GetApplicationRevisions(appGUID string) ([]Revision, Warnings, error) {
	var resources []Revision

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetApplicationRevisionsRequest,
		URIParams:    internal.Params{"app_guid": appGUID},
		ResponseBody: Revision{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Revision))
			return nil
		},
	})
	return resources, warnings, err
}
